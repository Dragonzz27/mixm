import uuid

from sqlalchemy import Column, Integer, String, ForeignKey, DateTime, func
from app.database import Base


class ContentCategory(Base):
    __tablename__ = "content_categories"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="绑定ID(UUID)",
    )
    content_id = Column(String(36), ForeignKey("contents.id", ondelete="CASCADE"), nullable=False, index=True, comment="内容ID")
    category_id = Column(String(36), ForeignKey("categories.id", ondelete="CASCADE"), nullable=False, index=True, comment="分类ID")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
