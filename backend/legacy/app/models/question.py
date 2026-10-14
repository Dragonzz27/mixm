import uuid

from sqlalchemy import (
    Column,
    String,
    Text,
    Integer,
    Boolean,
    DateTime,
    func,
    ForeignKey,
    JSON,
)
from sqlalchemy.orm import relationship
from app.database import Base
from app.models.relations import question_roles, question_tags


class Question(Base):
    __tablename__ = "questions"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="题目 ID(UUID)",
    )
    title = Column(
        String(500), nullable=False, index=True, comment="题目内容/题干"
    )
    type = Column(
        String(20),
        nullable=False,
        index=True,
        comment="题型: single_choice/multi_choice/true_false/blank/essay",
    )
    content = Column(Text, nullable=True, comment="详细解析/内容")
    answer = Column(Text, nullable=True, comment="答案")
    explanation = Column(Text, nullable=True, comment="解析")
    difficulty = Column(Integer, nullable=False, comment="难度 1-5")
    tags = Column(JSON, nullable=True, comment="简单标签列表 (非知识元)")
    source = Column(String(100), nullable=True, comment="来源")
    is_ai_generated = Column(
        Boolean, default=False, comment="是否AI生成"
    )
    created_by = Column(String(36), nullable=True, comment="创建者ID")
    review_status = Column(
        String(20),
        nullable=False,
        default="pending",
        server_default="pending",
        index=True,
        comment="review status: pending/approved/rejected",
    )
    reviewed_by = Column(String(36), nullable=True, comment="reviewer id")
    reviewed_at = Column(DateTime(timezone=True), nullable=True, comment="reviewed at")
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="创建时间"
    )
    updated_at = Column(
        DateTime(timezone=True), onupdate=func.now(), comment="更新时间"
    )

    visible_roles = relationship(
        "Role", secondary=question_roles
    )
    options = relationship(
        "QuestionOption", back_populates="question", cascade="all, delete-orphan"
    )
    attachments = relationship(
        "QuestionAttachment",
        back_populates="question",
        cascade="all, delete-orphan",
    )
    knowledge_tags = relationship(
        "Category",
        secondary=question_tags,
    )


class QuestionCategory(Base):
    __tablename__ = "question_categories"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="分类 ID(UUID)",
    )
    name = Column(String(100), nullable=False, comment="分类名称")
    parent_id = Column(
        String(36),
        ForeignKey("question_categories.id"),
        nullable=True,
        comment="父分类ID",
    )
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), comment="创建时间"
    )

    parent = relationship(
        "QuestionCategory", remote_side=[id], back_populates="children"
    )
    children = relationship(
        "QuestionCategory", back_populates="parent", cascade="all, delete-orphan"
    )


class QuestionOption(Base):
    __tablename__ = "question_options"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="选项 ID(UUID)",
    )
    question_id = Column(
        String(36),
        ForeignKey("questions.id"),
        nullable=False,
        comment="题目ID",
    )
    option_key = Column(
        String(10), nullable=False, comment="A/B/C/D"
    )
    content = Column(Text, nullable=False, comment="选项内容")
    is_correct = Column(Boolean, default=False, comment="是否正确答案")

    question = relationship("Question", back_populates="options")


class QuestionAttachment(Base):
    __tablename__ = "question_attachments"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="附件 ID(UUID)",
    )
    question_id = Column(
        String(36),
        ForeignKey("questions.id"),
        nullable=False,
        comment="题目ID",
    )
    file_path = Column(String(500), nullable=True, comment="文件存储路径")
    file_name = Column(String(200), nullable=True, comment="文件名称")
    file_type = Column(String(50), nullable=True, comment="文件类型")

    question = relationship("Question", back_populates="attachments")
