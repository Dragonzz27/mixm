from datetime import datetime
from typing import List

from pydantic import BaseModel, Field, field_validator


class RoleBase(BaseModel):
    name: str
    route_keys: List[str] = Field(default_factory=list)

    @field_validator("name")
    @classmethod
    def validate_name(cls, v):
        value = (v or "").strip()
        if not value:
            raise ValueError("Role name is required")
        return value


class RoleCreate(RoleBase):
    pass


class RoleUpdate(BaseModel):
    name: str
    route_keys: List[str] = Field(default_factory=list)

    @field_validator("name")
    @classmethod
    def validate_name(cls, v):
        value = (v or "").strip()
        if not value:
            raise ValueError("Role name is required")
        return value


class RoleResponse(RoleBase):
    id: str
    created_at: datetime
    updated_at: datetime | None = None

    class Config:
        from_attributes = True


class RoleListResponse(BaseModel):
    total: int
    roles: List[RoleResponse]


class PermissionOption(BaseModel):
    key: str
    label: str
