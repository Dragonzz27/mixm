import uuid

from sqlalchemy import Column, DateTime, ForeignKey, Integer, JSON, String, Table, Text, func
from sqlalchemy.orm import relationship

from app.database import Base


practice_roles = Table(
    "practice_roles",
    Base.metadata,
    Column("practice_id", String(36), ForeignKey("practices.id"), primary_key=True),
    Column("role_id", String(36), ForeignKey("roles.id"), primary_key=True),
)

practice_tags = Table(
    "practice_tags",
    Base.metadata,
    Column("practice_id", String(36), ForeignKey("practices.id"), primary_key=True),
    Column("category_id", String(36), ForeignKey("categories.id"), primary_key=True),
)


class Practice(Base):
    __tablename__ = "practices"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="practice id",
    )
    title = Column(String(200), nullable=False, index=True, comment="practice title")
    description = Column(Text, nullable=True, comment="practice description")
    type_counts = Column(JSON, nullable=True, comment="question type counts")
    created_by = Column(String(36), nullable=False, comment="creator user id")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="created at")
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), comment="updated at")

    visible_roles = relationship("Role", secondary=practice_roles)
    knowledge_tags = relationship("Category", secondary=practice_tags)
    practice_questions = relationship(
        "PracticeQuestion",
        back_populates="practice",
        cascade="all, delete-orphan",
        order_by="PracticeQuestion.sort_order",
    )


class PracticeQuestion(Base):
    __tablename__ = "practice_questions"

    id = Column(
        String(36),
        primary_key=True,
        index=True,
        default=lambda: str(uuid.uuid4()),
        comment="practice question id",
    )
    practice_id = Column(
        String(36),
        ForeignKey("practices.id"),
        nullable=False,
        index=True,
        comment="practice id",
    )
    question_id = Column(
        String(36),
        ForeignKey("questions.id"),
        nullable=False,
        index=True,
        comment="question id",
    )
    sort_order = Column(Integer, nullable=False, default=0, comment="sort order")

    practice = relationship("Practice", back_populates="practice_questions")
    question = relationship("Question")
