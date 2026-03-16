import uuid
from datetime import datetime
from typing import BinaryIO

import boto3
from botocore.config import Config

from app.config import get_settings


class StorageService:
    def __init__(self):
        settings = get_settings()
        self.endpoint = settings.RUSTFS_ENDPOINT
        self.bucket = settings.RUSTFS_BUCKET
        self.s3_client = boto3.client(
            "s3",
            endpoint_url=self.endpoint,
            aws_access_key_id=settings.RUSTFS_ACCESS_KEY,
            aws_secret_access_key=settings.RUSTFS_SECRET_KEY,
            region_name="us-east-1",
            config=Config(signature_version="s3v4"),
        )

    def upload_file(
        self, file: BinaryIO, file_name: str = None, content_type: str = None
    ) -> dict:
        ext = file_name.split(".")[-1] if file_name and "." in file_name else ""
        key = f"content/{datetime.now().strftime('%Y%m%d')}/{uuid.uuid4()}.{ext}"

        self.s3_client.upload_fileobj(
            file,
            self.bucket,
            key,
            ExtraArgs={"ContentType": content_type, "ACL": "public-read"},
        )

        return {"key": key, "content_type": content_type}

    def get_file(self, key: str) -> tuple:
        response = self.s3_client.get_object(Bucket=self.bucket, Key=key)
        content = response["Body"].read()
        content_type = response.get("ContentType", "application/octet-stream")
        return content, content_type


storage_service = StorageService()
