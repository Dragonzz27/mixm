from contextlib import asynccontextmanager
import os

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from sqlalchemy import inspect, text
from app.config import get_settings

from app.router import (
    auth,
    category,
    content,
    content_category,
    content_type,
    template,
    system,
    cluster,
    instance,
    image,
    role,
    assignment,
    question,
    knowledge_tree,
    question_generate,
)
from app.database import engine
from app.models import user
from app.models.category import Category as CategoryModel
from app.models.content import Content as ContentModel
from app.models.content_category import ContentCategory as ContentCategoryModel
from app.models.content_type import ContentType as ContentTypeModel
from app.models.template import Template as TemplateModel
from app.models.instance import Instance as InstanceModel
from app.models.image import ImageRepository
from app.models.role import Role as RoleModel
from app.models.user_role import UserRole as UserRoleModel
from app.models.assignment import Assignment, Submission, assignment_roles
from app.models.relations import question_roles, question_tags
from app.models.question import (
    Question,
    QuestionCategory,
    QuestionOption,
    QuestionAttachment,
)
from app.service.ssh_gateway import SSHGatewayService


def ensure_question_review_columns():
    inspector = inspect(engine)
    if "questions" not in inspector.get_table_names():
        return

    existing_columns = {column["name"] for column in inspector.get_columns("questions")}
    statements = []

    if "review_status" not in existing_columns:
        statements.append(
            "ALTER TABLE questions ADD COLUMN review_status VARCHAR(20) NOT NULL DEFAULT 'approved'"
        )
    if "reviewed_by" not in existing_columns:
        statements.append(
            "ALTER TABLE questions ADD COLUMN reviewed_by VARCHAR(36)"
        )
    if "reviewed_at" not in existing_columns:
        statements.append(
            "ALTER TABLE questions ADD COLUMN reviewed_at DATETIME"
        )
    if statements:
        with engine.begin() as conn:
            for statement in statements:
                conn.execute(text(statement))


user.Base.metadata.create_all(bind=engine)
CategoryModel.__table__.create(bind=engine, checkfirst=True)
ContentModel.__table__.create(bind=engine, checkfirst=True)
ContentCategoryModel.__table__.create(bind=engine, checkfirst=True)
ContentTypeModel.__table__.create(bind=engine, checkfirst=True)
TemplateModel.__table__.create(bind=engine, checkfirst=True)
InstanceModel.__table__.create(bind=engine, checkfirst=True)
ImageRepository.__table__.create(bind=engine, checkfirst=True)
RoleModel.__table__.create(bind=engine, checkfirst=True)
UserRoleModel.__table__.create(bind=engine, checkfirst=True)
assignment_roles.create(bind=engine, checkfirst=True)
Assignment.__table__.create(bind=engine, checkfirst=True)
Submission.__table__.create(bind=engine, checkfirst=True)
question_roles.create(bind=engine, checkfirst=True)
question_tags.create(bind=engine, checkfirst=True)
Question.__table__.create(bind=engine, checkfirst=True)
QuestionCategory.__table__.create(bind=engine, checkfirst=True)
QuestionOption.__table__.create(bind=engine, checkfirst=True)
QuestionAttachment.__table__.create(bind=engine, checkfirst=True)
ensure_question_review_columns()

ssh_gateway_service = SSHGatewayService()


@asynccontextmanager
async def lifespan(app: FastAPI):
    ssh_gateway_service.start()
    try:
        yield
    finally:
        ssh_gateway_service.stop()


app = FastAPI(lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

os.makedirs("static/export", exist_ok=True)
app.mount(
    "/static/export", StaticFiles(directory="static/export"), name="static/export"
)

app.include_router(auth.router, prefix="/api/auth", tags=["auth"])
app.include_router(category.router, prefix="/api/categories", tags=["categories"])
app.include_router(content.router, prefix="/api/contents", tags=["contents"])
app.include_router(content_category.router, prefix="/api/bindings", tags=["bindings"])
app.include_router(
    content_type.router, prefix="/api/content-types", tags=["content-types"]
)
app.include_router(template.router, prefix="/api/templates", tags=["templates"])
app.include_router(system.router, prefix="/api/system", tags=["system"])
app.include_router(cluster.router, prefix="/api/clusters", tags=["clusters"])
app.include_router(instance.router, prefix="/api/instances", tags=["instances"])
app.include_router(image.router, prefix="/api/images", tags=["images"])
app.include_router(role.router, prefix="/api/roles", tags=["roles"])
app.include_router(assignment.router, prefix="/api/assignments", tags=["assignments"])
app.include_router(question.router, prefix="/api/questions", tags=["questions"])
app.include_router(question.category_router, prefix="/api/question-categories", tags=["question-categories"])
app.include_router(knowledge_tree.router, prefix="/api/knowledge-tree", tags=["knowledge-tree"])
app.include_router(question_generate.router, prefix="/api/questions", tags=["questions"])
