import uuid

from sqlalchemy import (
    Column,
    String,
    DateTime,
    Boolean,
    func,
    ForeignKey,
)
from app.database import Base


class ImageRepository(Base):
    __tablename__ = "image_repositories"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="镜像仓库ID(UUID)",
    )
    name = Column(
        String(100), unique=True, nullable=False, index=True, comment="仓库名称"
    )
    is_public = Column(Boolean, default=False, comment="是否公开")
    user_id = Column(
        String(36), ForeignKey("users.id"), nullable=False, comment="创建者ID"
    )

    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="创建时间"
    )
    updated_at = Column(
        DateTime(timezone=True), onupdate=func.now(), comment="更新时间"
    )
