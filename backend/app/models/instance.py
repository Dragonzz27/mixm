from sqlalchemy import Column, Integer, String, Text, DateTime, func, ForeignKey
from app.database import Base


class Instance(Base):
    __tablename__ = "instances"

    id = Column(Integer, primary_key=True, index=True, comment="实例ID")
    name = Column(String(100), nullable=False, comment="实例名称")
    template_id = Column(Integer, ForeignKey("templates.id"), nullable=False, comment="模板ID")
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, comment="用户ID")
    
    status = Column(String(20), default="pending", comment="状态: pending/running/stopped/failed")
    namespace = Column(String(100), default="default", comment="命名空间")
    replicas = Column(Integer, default=1, comment="副本数")
    
    cpu_usage = Column(Integer, default=0, comment="CPU使用率(%)")
    memory_usage = Column(Integer, default=0, comment="内存使用(MB)")
    memory_limit = Column(Integer, default=1024, comment="内存限制(MB)")
    
    config = Column(Text, nullable=True, comment="实例配置")
    
    started_at = Column(DateTime(timezone=True), nullable=True, comment="启动时间")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")
