import uuid

from sqlalchemy import Column, Integer, String, Text, Boolean, DateTime, ForeignKey, func
from sqlalchemy.orm import relationship
from app.database import Base


class Category(Base):
    __tablename__ = "categories"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="分类ID(UUID)",
    )
    name = Column(String(100), nullable=False, index=True, comment="分类名称")
    description = Column(Text, nullable=True, comment="分类描述")
    parent_id = Column(String(36), ForeignKey("categories.id"), nullable=True, index=True, comment="父级分类ID")
    level = Column(Integer, default=1, index=True, comment="分类层级")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    is_active = Column(Boolean, default=True, comment="是否启用")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")

    parent = relationship("Category", remote_side=[id], backref="children")