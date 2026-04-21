from pydantic import BaseModel
from typing import Optional
from datetime import datetime


class TemplateBase(BaseModel):
    name: str
    description: Optional[str] = None
    config: Optional[str] = None


class TemplateCreate(TemplateBase):
    pass


class TemplateUpdate(TemplateBase):
    pass


class TemplateResponse(TemplateBase):
    id: str
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class TemplateListResponse(BaseModel):
    total: int
    templates: list[TemplateResponse]


class TemplateSearchParams(BaseModel):
    keyword: Optional[str] = None
