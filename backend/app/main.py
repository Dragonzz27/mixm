from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
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
from app.models.assignment import Assignment, Submission
from app.service.ssh_gateway import SSHGatewayService

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
Assignment.__table__.create(bind=engine, checkfirst=True)
Submission.__table__.create(bind=engine, checkfirst=True)

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
