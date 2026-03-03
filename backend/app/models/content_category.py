from sqlalchemy import Column, Integer, ForeignKey, DateTime, func
from app.database import Base


class ContentCategory(Base):
    __tablename__ = "content_categories"

    id = Column(Integer, primary_key=True, index=True, comment="绑定ID")
    content_id = Column(Integer, ForeignKey("contents.id", ondelete="CASCADE"), nullable=False, index=True, comment="内容ID")
    category_id = Column(Integer, ForeignKey("categories.id", ondelete="CASCADE"), nullable=False, index=True, comment="分类ID")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
