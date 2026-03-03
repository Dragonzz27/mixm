from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime


class ContentTypeBase(BaseModel):
    name: str
    code: str
    description: Optional[str] = None
    is_active: Optional[bool] = True


class ContentTypeCreate(ContentTypeBase):
    pass


class ContentTypeUpdate(ContentTypeBase):
    pass


class ContentTypeResponse(ContentTypeBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ContentTypeListResponse(BaseModel):
    total: int
    types: List[ContentTypeResponse]
