from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime


class ContentCategoryBase(BaseModel):
    content_id: str
    category_id: str
    sort_order: Optional[int] = 0


class ContentCategoryCreate(ContentCategoryBase):
    pass


class ContentCategoryResponse(ContentCategoryBase):
    id: str
    created_at: datetime

    class Config:
        from_attributes = True


class ContentWithCategories(BaseModel):
    id: str
    title: str
    content: Optional[str] = None
    content_type: str
    url: Optional[str] = None
    status: str
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None
    categories: List[dict] = []

    class Config:
        from_attributes = True


class CategoryWithContents(BaseModel):
    id: str
    name: str
    description: Optional[str] = None
    parent_id: Optional[str] = None
    level: int
    sort_order: int
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None
    contents: List[dict] = []

    class Config:
        from_attributes = True
