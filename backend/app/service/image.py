import requests
import logging
import time
from datetime import datetime
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import or_
from urllib.parse import quote

from app.models.image import ImageRepository
from app.schema.image import (
    ImageRepositoryCreate,
    ImageRepositoryUpdate,
    ImageTag,
)
from app.config import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


class RegistryService:
    DEFAULT_TIMEOUT_SECONDS = 10

    def __init__(self):
        self.base_url = getattr(settings, "REGISTRY_URL", "http://localhost:5000")
        self.username = getattr(settings, "REGISTRY_USERNAME", "")
        self.password = getattr(settings, "REGISTRY_PASSWORD", "")
        self._auth_headers = self._build_auth_headers()

    def _build_auth_headers(self) -> dict:
        if self.username and self.password:
            import base64

            auth = base64.b64encode(
                f"{self.username}:{self.password}".encode()
            ).decode()
            return {"Authorization": f"Basic {auth}"}
        return {}

    def _build_manifest_url(self, repository_name: str, reference: str) -> str:
        encoded_repo = quote(repository_name, safe="/")
        encoded_ref = quote(reference, safe="")
        return f"{self.base_url}/v2/{encoded_repo}/manifests/{encoded_ref}"

    def _build_tags_url(self, repository_name: str) -> str:
        encoded_repo = quote(repository_name, safe="/")
        return f"{self.base_url}/v2/{encoded_repo}/tags/list"

    def _request(self, method: str, url: str, headers: Optional[dict] = None):
        req_headers = self._auth_headers.copy()
        if headers:
            req_headers.update(headers)
        return requests.request(
            method,
            url,
            headers=req_headers,
            timeout=self.DEFAULT_TIMEOUT_SECONDS,
            verify=False,
        )

    def _fetch_manifest(self, repository_name: str, reference: str) -> Optional[dict]:
        url = self._build_manifest_url(repository_name, reference)
        accept_headers = [
            "application/vnd.oci.image.manifest.v1+json",
            "application/vnd.docker.distribution.manifest.v2+json",
            "application/vnd.oci.image.index.v1+json",
            "application/vnd.docker.distribution.manifest.list.v2+json",
        ]

        for accept in accept_headers:
            response = self._request("GET", url, headers={"Accept": accept})
            if response.status_code == 200:
                return response.json()

        return None

    def _calculate_manifest_size(self, repository_name: str, manifest: dict) -> int:
        if not isinstance(manifest, dict):
            return 0

        if "layers" in manifest:
            config = manifest.get("config", {})
            config_size = config.get("size", 0) if isinstance(config, dict) else 0
            layers_size = sum(
                layer.get("size", 0)
                for layer in manifest.get("layers", [])
                if isinstance(layer, dict)
            )
            return config_size + layers_size

        if "manifests" in manifest:
            total_size = 0
            fallback_size = 0

            for sub in manifest.get("manifests", []):
                if not isinstance(sub, dict):
                    continue

                fallback_size += sub.get("size", 0)
                digest = sub.get("digest")
                if not digest:
                    continue

                sub_manifest = self._fetch_manifest(repository_name, digest)
                if sub_manifest:
                    total_size += self._calculate_manifest_size(
                        repository_name, sub_manifest
                    )

            return total_size or fallback_size

        return 0

    def get_tags(self, repository_name: str) -> List[ImageTag]:
        try:
            url = self._build_tags_url(repository_name)
            response = self._request("GET", url)

            if response.status_code == 404:
                logger.warning(f"Repository {repository_name} not found in registry")
                return []

            if response.status_code != 200:
                logger.error(
                    f"Failed to get tags for {repository_name}: {response.status_code}"
                )
                return []

            data = response.json()
            tags = data.get("tags") or []

            if not tags:
                return []

            result = [
                self._get_tag_manifest(repository_name, tag_name) for tag_name in tags
            ]

            return result

        except Exception as e:
            logger.error(f"Error getting tags from registry: {str(e)}")
            return []

    def _get_tag_manifest(self, repository_name: str, tag_name: str) -> ImageTag:
        try:
            manifest = self._fetch_manifest(repository_name, tag_name)
            response_status = 200 if manifest else 404

            if manifest:
                size = self._calculate_manifest_size(repository_name, manifest)

                created = manifest.get("created")
                pushed_at = None
                if created:
                    try:
                        pushed_at = datetime.fromisoformat(
                            created.replace("Z", "+00:00")
                        )
                    except (TypeError, ValueError):
                        pass

                return ImageTag(
                    name=tag_name, size=size, pushed_at=pushed_at, created_at=pushed_at
                )
            else:
                logger.warning(
                    f"Failed to get manifest for {repository_name}:{tag_name}, status: {response_status}"
                )

        except Exception as e:
            logger.error(
                f"Error getting manifest for {repository_name}:{tag_name}: {str(e)}"
            )

        return ImageTag(name=tag_name, size=0, pushed_at=None, created_at=None)

    def delete_tag(self, repository_name: str, tag_name: str) -> bool:
        try:
            url = self._build_manifest_url(repository_name, tag_name)
            accept_headers = [
                "application/vnd.oci.image.manifest.v1+json",
                "application/vnd.docker.distribution.manifest.v2+json",
            ]

            for accept in accept_headers:
                headers = {"Accept": accept}
                response = self._request("GET", url, headers=headers)

                if response.status_code == 404:
                    continue

                if response.status_code == 200:
                    digest = response.headers.get("Docker-Content-Digest")
                    if digest:
                        delete_url = self._build_manifest_url(repository_name, digest)
                        del_response = self._request(
                            "DELETE", delete_url, headers=headers
                        )
                        return del_response.status_code in [200, 202]

                logger.warning(
                    f"Unexpected status when deleting tag {repository_name}:{tag_name}: {response.status_code}"
                )
                return False

            return True

        except Exception as e:
            logger.error(f"Error deleting tag from registry: {str(e)}")
            return False

    def tag_exists(self, repository_name: str, tag_name: str) -> Optional[bool]:
        try:
            url = self._build_manifest_url(repository_name, tag_name)
            accept_headers = [
                "application/vnd.oci.image.manifest.v1+json",
                "application/vnd.docker.distribution.manifest.v2+json",
            ]

            for accept in accept_headers:
                response = self._request("GET", url, headers={"Accept": accept})

                if response.status_code == 200:
                    return True

                if response.status_code == 404:
                    continue

                logger.warning(
                    f"Unexpected status when checking tag existence for {repository_name}:{tag_name}: {response.status_code}"
                )
                return None

            return False

        except Exception as e:
            logger.error(
                f"Error checking tag existence for {repository_name}:{tag_name}: {str(e)}"
            )
            return None

    def wait_for_tags_deleted(
        self,
        repository_name: str,
        tag_names: List[str],
        timeout_seconds: int = 30,
        poll_interval_seconds: int = 1,
    ) -> bool:
        if not tag_names:
            return True

        deadline = time.monotonic() + timeout_seconds
        remaining_tags = set(tag_names)

        while remaining_tags and time.monotonic() < deadline:
            not_deleted = set()
            for tag_name in remaining_tags:
                exists = self.tag_exists(repository_name, tag_name)
                if exists is True or exists is None:
                    not_deleted.add(tag_name)

            if not not_deleted:
                return True

            remaining_tags = not_deleted
            time.sleep(poll_interval_seconds)

        logger.error(
            f"Timeout waiting for tags to be deleted in {repository_name}, remaining: {sorted(list(remaining_tags))}"
        )
        return False


class ImageService:
    def __init__(self):
        self.registry_service = RegistryService()

    def create(
        self, db: Session, repo: ImageRepositoryCreate, user_id: int
    ) -> ImageRepository:
        existing = (
            db.query(ImageRepository).filter(ImageRepository.name == repo.name).first()
        )
        if existing:
            raise ValueError("仓库名称已存在")

        db_repo = ImageRepository(
            name=repo.name,
            is_public=repo.is_public,
            user_id=user_id,
        )
        db.add(db_repo)
        db.commit()
        db.refresh(db_repo)
        return db_repo

    def get_all(
        self, db: Session, skip: int = 0, limit: int = 100
    ) -> List[ImageRepository]:
        return db.query(ImageRepository).offset(skip).limit(limit).all()

    def get_by_user(self, db: Session, user_id: int) -> List[ImageRepository]:
        return (
            db.query(ImageRepository).filter(ImageRepository.user_id == user_id).all()
        )

    def get_public_and_user(self, db: Session, user_id: int) -> List[ImageRepository]:
        return (
            db.query(ImageRepository)
            .filter(
                or_(
                    ImageRepository.is_public.is_(True),
                    ImageRepository.user_id == user_id,
                )
            )
            .all()
        )

    def get_public(self, db: Session) -> List[ImageRepository]:
        return (
            db.query(ImageRepository).filter(ImageRepository.is_public.is_(True)).all()
        )

    def get_by_id(self, db: Session, repo_id: int) -> Optional[ImageRepository]:
        return db.query(ImageRepository).filter(ImageRepository.id == repo_id).first()

    def update(
        self, db: Session, repo_id: int, repo: ImageRepositoryUpdate
    ) -> Optional[ImageRepository]:
        db_repo = self.get_by_id(db, repo_id)
        if not db_repo:
            return None

        if repo.is_public is not None:
            db_repo.is_public = repo.is_public

        db.commit()
        db.refresh(db_repo)
        return db_repo

    def delete(self, db: Session, repo_id: int) -> bool:
        db_repo = self.get_by_id(db, repo_id)
        if not db_repo:
            return False

        tags = self.registry_service.get_tags(db_repo.name)
        tag_names = [tag.name for tag in tags]

        for tag_name in tag_names:
            deleted = self.registry_service.delete_tag(db_repo.name, tag_name)
            if not deleted:
                exists = self.registry_service.tag_exists(db_repo.name, tag_name)
                if exists is True or exists is None:
                    logger.error(
                        f"Failed to delete tag {db_repo.name}:{tag_name}, aborting repository delete"
                    )
                    return False

        if tag_names:
            timeout_seconds = getattr(settings, "REGISTRY_DELETE_TIMEOUT_SECONDS", 30)
            poll_interval_seconds = getattr(
                settings, "REGISTRY_DELETE_POLL_INTERVAL_SECONDS", 1
            )
            all_deleted = self.registry_service.wait_for_tags_deleted(
                db_repo.name,
                tag_names,
                timeout_seconds=timeout_seconds,
                poll_interval_seconds=poll_interval_seconds,
            )
            if not all_deleted:
                logger.error(
                    f"Registry delete not fully completed for repository {db_repo.name}, keeping database record"
                )
                return False

        db.delete(db_repo)
        db.commit()
        return True

    def get_tags(self, db: Session, repo_id: int) -> List[ImageTag]:
        db_repo = self.get_by_id(db, repo_id)
        if not db_repo:
            return []

        return self.registry_service.get_tags(db_repo.name)

    def delete_tag(self, db: Session, repo_id: int, tag_name: str) -> bool:
        db_repo = self.get_by_id(db, repo_id)
        if not db_repo:
            return False

        return self.registry_service.delete_tag(db_repo.name, tag_name)

    def can_access(self, db_repo: ImageRepository, user_id: int) -> bool:
        return db_repo.is_public or db_repo.user_id == user_id

    def can_modify(self, db_repo: ImageRepository, user_id: int) -> bool:
        return db_repo.user_id == user_id


image_service = ImageService()
