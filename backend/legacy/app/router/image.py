from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import List, Optional
import logging

from app.database import get_db
from app.models.user import User
from app.schema.image import (
    ImageRepositoryCreate,
    ImageRepositoryUpdate,
    ImageRepositoryResponse,
    ImageRepositoryListResponse,
    ImageTag,
    ImageRepositoryWithTags,
    ImageRepositoryWithUser,
    ImageCascaderResponse,
    ImageRepositoryOption,
    ImageTagOption,
)
from app.service.image import image_service
from app.service.auth import AuthService

router = APIRouter()
auth_service = AuthService()
logger = logging.getLogger(__name__)


def _build_repo_with_user_list(db: Session, repos) -> List[ImageRepositoryWithUser]:
    if not repos:
        return []

    owner_ids = {repo.user_id for repo in repos}
    users = db.query(User).filter(User.id.in_(owner_ids)).all()
    user_map = {user.id: user for user in users}

    result = []
    for repo in repos:
        tags = image_service.get_tags(db, repo.id)
        result.append(
            ImageRepositoryWithUser(
                id=repo.id,
                name=repo.name,
                is_public=repo.is_public,
                user_id=repo.user_id,
                created_at=repo.created_at,
                updated_at=repo.updated_at,
                tags_count=len(tags),
                user_info=user_map.get(repo.user_id),
            )
        )

    return result


@router.post(
    "/", response_model=ImageRepositoryResponse, status_code=status.HTTP_201_CREATED
)
def create_repository(
    repo: ImageRepositoryCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    try:
        result = image_service.create(db, repo, current_user.id)
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.exception("创建仓库失败")
        raise HTTPException(status_code=500, detail="创建仓库失败")


@router.get("/", response_model=ImageRepositoryListResponse)
def list_repositories(
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    repo_type: str = Query("my", alias="type", description="my:我的, public:公开"),
    search: Optional[str] = Query(None, description="搜索镜像名称"),
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    if repo_type not in {"my", "public"}:
        raise HTTPException(status_code=422, detail="type 参数仅支持 my 或 public")

    if repo_type == "my":
        repos = image_service.get_by_user(db, current_user.id)
    else:
        repos = image_service.get_public(db)

    if search:
        search_key = search.strip().lower()
        if search_key:
            repos = [r for r in repos if search_key in r.name.lower()]

    total = len(repos)
    repos = repos[skip : skip + limit]
    result = _build_repo_with_user_list(db, repos)

    return {
        "total": total,
        "repositories": result,
    }


@router.get("/me", response_model=ImageRepositoryListResponse)
def get_my_repositories(
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    search: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    repos = image_service.get_by_user(db, current_user.id)

    if search:
        search_key = search.strip().lower()
        if search_key:
            repos = [r for r in repos if search_key in r.name.lower()]

    total = len(repos)
    repos = repos[skip : skip + limit]
    result = _build_repo_with_user_list(db, repos)

    return {
        "total": total,
        "repositories": result,
    }


@router.get("/{repo_id}", response_model=ImageRepositoryWithTags)
def get_repository(
    repo_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    db_repo = image_service.get_by_id(db, repo_id)
    if not db_repo:
        raise HTTPException(status_code=404, detail="仓库不存在")

    if not image_service.can_access(db_repo, current_user.id):
        raise HTTPException(status_code=403, detail="无权访问此仓库")

    tags = image_service.get_tags(db, repo_id)
    user = db.query(User).filter(User.id == db_repo.user_id).first()

    return {
        "id": db_repo.id,
        "name": db_repo.name,
        "is_public": db_repo.is_public,
        "user_id": db_repo.user_id,
        "created_at": db_repo.created_at,
        "updated_at": db_repo.updated_at,
        "tags": tags,
        "tags_count": len(tags),
        "user_info": user,
    }


@router.put("/{repo_id}", response_model=ImageRepositoryResponse)
def update_repository(
    repo_id: str,
    repo: ImageRepositoryUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    db_repo = image_service.get_by_id(db, repo_id)
    if not db_repo:
        raise HTTPException(status_code=404, detail="仓库不存在")

    if not image_service.can_modify(db_repo, current_user.id):
        raise HTTPException(status_code=403, detail="无权修改此仓库")

    result = image_service.update(db, repo_id, repo)
    return result


@router.delete("/{repo_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_repository(
    repo_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    db_repo = image_service.get_by_id(db, repo_id)
    if not db_repo:
        raise HTTPException(status_code=404, detail="仓库不存在")

    if not image_service.can_modify(db_repo, current_user.id):
        raise HTTPException(status_code=403, detail="无权删除此仓库")

    success = image_service.delete(db, repo_id)
    if not success:
        raise HTTPException(status_code=409, detail="仓库镜像尚未删除完成，请稍后重试")


@router.get("/{repo_id}/tags", response_model=List[ImageTag])
def get_repository_tags(
    repo_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    db_repo = image_service.get_by_id(db, repo_id)
    if not db_repo:
        raise HTTPException(status_code=404, detail="仓库不存在")

    if not image_service.can_access(db_repo, current_user.id):
        raise HTTPException(status_code=403, detail="无权访问此仓库")

    return image_service.get_tags(db, repo_id)


@router.delete("/{repo_id}/tags/{tag_name}", status_code=status.HTTP_204_NO_CONTENT)
def delete_repository_tag(
    repo_id: str,
    tag_name: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(auth_service.get_current_user),
):
    tag_name = tag_name.strip()
    if not tag_name:
        raise HTTPException(status_code=422, detail="tag 名称不能为空")

    db_repo = image_service.get_by_id(db, repo_id)
    if not db_repo:
        raise HTTPException(status_code=404, detail="仓库不存在")

    if not image_service.can_modify(db_repo, current_user.id):
        raise HTTPException(status_code=403, detail="无权删除此镜像")

    success = image_service.delete_tag(db, repo_id, tag_name)
    if not success:
        raise HTTPException(status_code=500, detail="删除镜像失败")
