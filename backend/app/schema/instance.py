from pydantic import BaseModel, field_validator, model_validator
from typing import Optional, List
from datetime import datetime


class InstanceConfig(BaseModel):
    repo_id: int
    repo_tag: str
    cpu_limit: float = 1.0
    memory_limit: str = "1Gi"
    storage_limit: int = 1
    gpu_limit: int = 0
    gpu_memory_limit: int = 0

    @field_validator("cpu_limit", mode="before")
    @classmethod
    def validate_cpu(cls, v):
        if v < 0.5:
            raise ValueError("CPU不能小于0.5")
        if v > 16:
            raise ValueError("CPU不能大于16")
        return v

    @field_validator("storage_limit", mode="before")
    @classmethod
    def validate_storage(cls, v):
        if v < 1:
            raise ValueError("存储不能小于1Gi")
        if v > 40:
            raise ValueError("存储不能大于40Gi")
        return v

    @field_validator("gpu_limit", mode="before")
    @classmethod
    def validate_gpu(cls, v):
        if v < 0:
            raise ValueError("GPU数量不能小于0")
        if v > 8:
            raise ValueError("GPU数量不能大于8")
        return v


    @field_validator("gpu_memory_limit", mode="before")
    @classmethod
    def validate_gpu_memory(cls, v):
        if v < 0:
            raise ValueError("GPU显存不能小于0MB")
        # 显存配置暂不做上限限制，因为不同型号GPU显存差异较大
        return v

    @model_validator(mode="after")
    def validate_memory(self):
        memory_str = self.memory_limit.upper().strip()
        if memory_str.endswith("GI"):
            memory_mb = int(memory_str[:-2]) * 1024
        elif memory_str.endswith("G"):
            memory_mb = int(memory_str[:-1]) * 1024
        elif memory_str.endswith("MI"):
            memory_mb = int(memory_str[:-2])
        elif memory_str.endswith("M"):
            memory_mb = int(memory_str[:-1])
        else:
            raise ValueError("内存格式不正确，请使用Gi或Mi为单位")

        if memory_mb < 512:
            raise ValueError("内存不能小于512Mi")
        if memory_mb > 32 * 1024:
            raise ValueError("内存不能大于32Gi")
        return self


class InstanceBase(BaseModel):
    name: str
    user_id: int
    image: Optional[str] = "ubuntu:24.04"
    status: Optional[str] = "pending"
    namespace: Optional[str] = "default"
    replicas: Optional[int] = 1
    cpu_usage: Optional[int] = 0
    memory_usage: Optional[int] = 0
    memory_limit: Optional[str] = "1Gi"
    memory_limit_mb: Optional[int] = 1024
    storage_limit: Optional[int] = 1
    cpu_limit: Optional[float] = 1.0
    repo_id: Optional[int] = None
    repo_tag: Optional[str] = None
    ssh_port: Optional[int] = None
    ssh_username: Optional[str] = None
    config: Optional[str] = None


class InstanceCreate(InstanceBase):
    pass


class InstanceCreateWithConfig(BaseModel):
    name: str = ""
    user_id: Optional[int] = None
    repo_id: int
    repo_tag: str
    cpu_limit: float = 1.0
    memory_limit: str = "1Gi"
    storage_limit: int = 1
    gpu_limit: int = 0
    gpu_memory_limit: int = 0

    @field_validator("cpu_limit", mode="before")
    @classmethod
    def validate_cpu(cls, v):
        if v < 0.5:
            raise ValueError("CPU不能小于0.5")
        if v > 16:
            raise ValueError("CPU不能大于16")
        return v

    @field_validator("storage_limit", mode="before")
    @classmethod
    def validate_storage(cls, v):
        if v < 1:
            raise ValueError("存储不能小于1Gi")
        if v > 40:
            raise ValueError("存储不能大于40Gi")
        return v

    @field_validator("gpu_limit", mode="before")
    @classmethod
    def validate_gpu(cls, v):
        if v < 0:
            raise ValueError("GPU数量不能小于0")
        if v > 8:
            raise ValueError("GPU数量不能大于8")
        return v


class InstanceUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[str] = None
    replicas: Optional[int] = None
    cpu_usage: Optional[int] = None
    memory_usage: Optional[int] = None
    gpu_usage: Optional[int] = None
    gpu_memory_usage: Optional[int] = None
    

class InstanceResponse(InstanceBase):
    id: str
    username: Optional[str] = None
    endpoint: Optional[str] = None
    hostIp: Optional[str] = None
    port: Optional[int] = None
    ssh_password: Optional[str] = None
    ssh_connection: Optional[str] = None
    started_at: Optional[datetime] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class InstanceListResponse(BaseModel):
    total: int
    instances: list[InstanceResponse]
