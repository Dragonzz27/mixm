from pydantic import BaseModel
from typing import Optional
from datetime import datetime


class InstanceBase(BaseModel):
    name: str
    template_id: int
    user_id: int
    status: Optional[str] = "pending"
    namespace: Optional[str] = "default"
    replicas: Optional[int] = 1
    cpu_usage: Optional[int] = 0
    memory_usage: Optional[int] = 0
    memory_limit: Optional[int] = 1024
    config: Optional[str] = None


class InstanceCreate(InstanceBase):
    pass


class InstanceUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[str] = None
    replicas: Optional[int] = None
    cpu_usage: Optional[int] = None
    memory_usage: Optional[int] = None


class InstanceResponse(InstanceBase):
    id: int
    username: Optional[str] = None
    template_name: Optional[str] = None
    endpoint: Optional[str] = None
    hostIp: Optional[str] = None
    port: Optional[int] = None
    started_at: Optional[datetime] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class InstanceListResponse(BaseModel):
    total: int
    instances: list[InstanceResponse]
