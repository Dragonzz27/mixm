import uuid
from sqlalchemy import Column, Integer, String, Text, DateTime, Float, func, ForeignKey
from app.database import Base
from sqlalchemy.dialects.postgresql import UUID as PG_UUID


class Instance(Base):
    __tablename__ = "instances"

    id = Column(
        String(36),
        primary_key=True,
        default=lambda: str(uuid.uuid4()),
        index=True,
        comment="实例ID (UUID)",
    )

    name = Column(String(100), nullable=False, comment="实例名称")
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, comment="用户ID")
    image = Column(
        String(200), default="ubuntu:24.04", nullable=False, comment="容器镜像"
    )

    repo_id = Column(
        Integer,
        ForeignKey("image_repositories.id"),
        nullable=True,
        comment="镜像仓库ID",
    )
    repo_tag = Column(String(100), nullable=True, comment="镜像标签")

    status = Column(
        String(20), default="pending", comment="状态: pending/running/stopped/failed"
    )
    namespace = Column(String(100), default="default", comment="命名空间")
    replicas = Column(Integer, default=1, comment="副本数")

    cpu_limit = Column(Float, default=1.0, comment="CPU限制(核)")
    memory_limit = Column(String(20), default="1Gi", comment="内存限制")
    memory_limit_mb = Column(Integer, default=1024, comment="内存限制(MB)")
    storage_limit = Column(Integer, default=1, comment="容器存储限制(Gi)")

    gpu_limit = Column(Integer, default=0, comment="GPU限制(卡)")
    gpu_memory_limit = Column(Integer, default=0, comment="单张显存限制(MB)")

    cpu_usage = Column(Integer, default=0, comment="CPU使用率(%)")
    memory_usage = Column(Integer, default=0, comment="内存使用(MB)")

    ssh_port = Column(Integer, nullable=True, comment="SSH端口")
    ssh_username = Column(String(50), nullable=True, comment="SSH用户名")

    config = Column(Text, nullable=True, comment="实例配置")

    started_at = Column(DateTime(timezone=True), nullable=True, comment="启动时间")
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="创建时间"
    )
    updated_at = Column(
        DateTime(timezone=True), onupdate=func.now(), comment="更新时间"
    )
