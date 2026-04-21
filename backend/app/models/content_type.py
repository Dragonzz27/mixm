import uuid

from sqlalchemy import Column, String, Boolean, DateTime, func
from app.database import Base


class ContentType(Base):
    __tablename__ = "content_types"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="类型ID(UUID)",
    )
    name = Column(String(50), unique=True, index=True, nullable=False, comment="类型名称")
    code = Column(String(50), unique=True, index=True, nullable=False, comment="类型代码")
    description = Column(String(200), nullable=True, comment="类型描述")
    is_active = Column(Boolean, default=True, comment="是否启用")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")
