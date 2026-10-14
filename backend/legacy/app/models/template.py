import uuid

from sqlalchemy import Column, String, Text, DateTime, func
from app.database import Base


class Template(Base):
    __tablename__ = "templates"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="模板ID(UUID)",
    )
    name = Column(String(100), unique=True, index=True, nullable=False, comment="模板名称")
    description = Column(String(500), nullable=True, comment="模板描述")
    config = Column(Text, nullable=True, comment="模板配置")

    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")
