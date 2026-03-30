from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime


class ImageRepositoryBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100, description="仓库名称")
    is_public: bool = Field(default=False, description="是否公开")


class ImageRepositoryCreate(ImageRepositoryBase):
    pass


class ImageRepositoryUpdate(BaseModel):
    is_public: Optional[bool] = None


class ImageTag(BaseModel):
    name: str = Field(..., description="标签名称")
    size: int = Field(0, description="镜像大小(字节)")
    pushed_at: Optional[datetime] = None
    created_at: Optional[datetime] = None


class UserInfo(BaseModel):
    id: str
    username: str

    class Config:
        from_attributes = True


class ImageRepositoryResponse(ImageRepositoryBase):
    id: int
    user_id: str
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ImageRepositoryWithUser(ImageRepositoryResponse):
    tags_count: int = Field(0, description="镜像标签数量")
    user_info: Optional[UserInfo] = None


class ImageRepositoryWithTags(ImageRepositoryResponse):
    tags: List[ImageTag] = Field(default_factory=list, description="镜像标签列表")
    tags_count: int = Field(0, description="标签数量")
    user_info: Optional[UserInfo] = None


class ImageRepositoryListResponse(BaseModel):
    total: int
    repositories: List[ImageRepositoryWithUser]


class ImageTagOption(BaseModel):
    value: str = Field(..., description="标签名称")
    label: str = Field(..., description="标签名称")


class ImageRepositoryOption(BaseModel):
    value: int = Field(..., description="仓库ID")
    label: str = Field(..., description="仓库名称")
    children: Optional[List[ImageTagOption]] = Field(
        default_factory=list, description="标签列表"
    )


class ImageCascaderResponse(BaseModel):
    repositories: List[ImageRepositoryOption] = Field(
        default_factory=list, description="镜像仓库级联选择数据"
    )
