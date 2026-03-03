from sqlalchemy import Column, Integer, String, Text, Boolean, DateTime, func, ForeignKey
from app.database import Base


class Content(Base):
    __tablename__ = "contents"

    id = Column(Integer, primary_key=True, index=True, comment="内容ID")
    title = Column(String(200), nullable=False, index=True, comment="标题")
    content = Column(Text, nullable=True, comment="内容详情")
    content_type = Column(String(50), default="article", comment="内容类型: article/video/exercise等")
    url = Column(String(500), nullable=True, comment="资源链接")
    status = Column(String(20), default="draft", index=True, comment="状态: draft/published/archived")
    template_id = Column(Integer, ForeignKey("templates.id"), nullable=True, comment="模板ID")
    is_active = Column(Boolean, default=True, comment="是否启用")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")
