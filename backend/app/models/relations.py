from sqlalchemy import Table, Column, String, ForeignKey
from app.database import Base

question_roles = Table(
    "question_roles",
    Base.metadata,
    Column(
        "question_id",
        String(36),
        ForeignKey("questions.id"),
        primary_key=True,
    ),
    Column("role_id", String(36), ForeignKey("roles.id"), primary_key=True),
)

question_tags = Table(
    "question_tags",
    Base.metadata,
    Column(
        "question_id",
        String(36),
        ForeignKey("questions.id"),
        primary_key=True,
    ),
    Column(
        "category_id",
        String(36),
        ForeignKey("categories.id"),
        primary_key=True,
        comment="知识元/知识点分类ID",
    ),
)