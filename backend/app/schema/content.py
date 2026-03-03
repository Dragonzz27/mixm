from pydantic import BaseModel, field_validator
from typing import Optional
from datetime import datetime


class ContentBase(BaseModel):
    title: str
    content: Optional[str] = None
    content_type: Optional[str] = "article"
    url: Optional[str] = None
    status: Optional[str] = "draft"
    template_id: Optional[int] = None
    is_active: Optional[bool] = True

    @field_validator("content_type", "status", mode="before")
    @classmethod
    def empty_to_none(cls, v):
        if v == "":
            return None
        return v

    @field_validator("template_id", mode="before")
    @classmethod
    def template_id_empty_to_none(cls, v):
        if v == "" or v is None:
            return None
        return v


class ContentCreate(ContentBase):
    pass


class ContentUpdate(ContentBase):
    pass


class ContentBriefResponse(BaseModel):
    id: int
    title: str
    content_type: str
    url: Optional[str] = None
    status: str
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ContentResponse(ContentBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ContentListResponse(BaseModel):
    total: int
    contents: list[ContentBriefResponse]
