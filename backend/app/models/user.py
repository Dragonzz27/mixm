import enum
import uuid

from sqlalchemy import Column, Integer, String, Boolean, DateTime, func, JSON, Enum
from app.database import Base


class User(Base):
    __tablename__ = "users"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="用户ID(UUID)",
    )
    username = Column(String(50), unique=True, index=True, nullable=False, comment="用户名")
    email = Column(String(100), unique=True, index=True, nullable=True, comment="邮箱地址")
    phone = Column(String(20), unique=True, index=True, nullable=True, comment="手机号码")
    hashed_password = Column(String(255), nullable=False, comment="加密后的密码")
    is_active = Column(Boolean, default=True, comment="是否激活")
    is_superuser = Column(Boolean, default=False, comment="是否超级管理员")
    
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="更新时间")

class PolicyEffect(enum.Enum):
    ALLOW = "allow"
    DENY = "deny"

class Policy(Base):
    __tablename__ = "policies"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
    )
    name = Column(String(100), unique=True, nullable=False, comment="策略名称")
    description = Column(String(255), comment="描述")
    
    # 核心 ABAC 规则定义
    # condition 格式示例: {"subject.department": "HR", "resource.type": "salary", "action": "read"}
    condition = Column(JSON, nullable=False, comment="策略判定规则（主体、资源、环境属性的组合）")
    
    action = Column(String(50), nullable=False, comment="操作类型: read, write, delete, etc.")
    resource_type = Column(String(50), index=True, nullable=False, comment="资源类型")
    
    effect = Column(Enum(PolicyEffect), default=PolicyEffect.ALLOW, comment="效力: 允许或拒绝")
    priority = Column(Integer, default=0, comment="优先级")
    is_active = Column(Boolean, default=True, comment="是否启用")
