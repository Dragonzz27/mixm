from pydantic import BaseModel, field_validator
from typing import Optional, List
from datetime import datetime


class CategoryBase(BaseModel):
    name: str
    description: Optional[str] = None
    parent_id: Optional[str] = None
    level: Optional[int] = 1
    sort_order: Optional[int] = 0
    is_active: Optional[bool] = True


class CategoryCreate(CategoryBase):
    @field_validator('parent_id')
    @classmethod
    def validate_parent_id(cls, v, info):
        if v is not None and v == info.data.get('id'):
            raise ValueError('Category cannot be its own parent')
        return v


class CategoryUpdate(CategoryBase):
    pass


class CategoryResponse(CategoryBase):
    id: str
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class CategoryTreeResponse(CategoryBase):
    id: str
    created_at: datetime
    updated_at: Optional[datetime] = None
    children: List['CategoryTreeResponse'] = []

    class Config:
        from_attributes = True


CategoryTreeResponse.model_rebuild()
