import uuid

from sqlalchemy import Column, DateTime, JSON, String, func

from app.database import Base


class Role(Base):
    __tablename__ = "roles"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="角色ID(UUID)",
    )
    name = Column(String(100), unique=True, index=True, nullable=False, comment="角色名")
    route_keys = Column(JSON, nullable=False, default=list, comment="路由权限key列表")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")
