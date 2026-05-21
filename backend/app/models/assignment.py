import uuid

from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    DateTime,
    func,
    ForeignKey,
    Table,
)
from sqlalchemy.orm import relationship
from app.database import Base


assignment_roles = Table(
    "assignment_roles",
    Base.metadata,
    Column(
        "assignment_id",
        String(36),
        ForeignKey("assignments.id"),
        primary_key=True,
    ),
    Column("role_id", String(36), ForeignKey("roles.id"), primary_key=True),
)


class Assignment(Base):
    __tablename__ = "assignments"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="作业 ID(UUID)",
    )
    title = Column(String(200), nullable=False, index=True, comment="作业标题")
    description = Column(Text, nullable=True, comment="作业描述")
    attachments = Column(Text, nullable=True, comment="作业附件 JSON 列表")
    content_id = Column(
        String(36),
        ForeignKey("contents.id"),
        nullable=True,
        comment="关联内容 ID",
    )
    deadline = Column(DateTime(timezone=True), nullable=False, comment="截止时间")
    status = Column(
        String(20),
        default="active",
        index=True,
        comment="状态：active/completed/archived",
    )
    created_by = Column(String(36), nullable=False, comment="创建者用户 ID")
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="创建时间"
    )
    updated_at = Column(
        DateTime(timezone=True), onupdate=func.now(), comment="更新时间"
    )

    visible_roles = relationship(
        "Role", secondary=assignment_roles, back_populates="assignments"
    )


class Submission(Base):
    __tablename__ = "submissions"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="提交 ID(UUID)",
    )
    assignment_id = Column(
        String(36),
        ForeignKey("assignments.id"),
        nullable=False,
        comment="作业 ID",
    )
    user_id = Column(
        String(36), ForeignKey("users.id"), nullable=False, comment="提交者用户 ID"
    )
    file_path = Column(String(500), nullable=True, comment="文件存储路径")
    file_name = Column(String(200), nullable=True, comment="文件名称")
    file_size = Column(Integer, nullable=True, comment="文件大小 (bytes)")
    content = Column(Text, nullable=True, comment="提交内容（如果是文本）")
    status = Column(
        String(20),
        default="submitted",
        index=True,
        comment="状态：submitted/reviewed/completed",
    )
    grade = Column(Integer, nullable=True, comment="评分")
    feedback = Column(Text, nullable=True, comment="反馈")
    submitted_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="提交时间"
    )
    reviewed_at = Column(DateTime(timezone=True), nullable=True, comment="审核时间")
    updated_at = Column(
        DateTime(timezone=True), onupdate=func.now(), comment="更新时间"
    )

    user = relationship("User", foreign_keys=[user_id])
