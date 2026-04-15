from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field, field_validator


class AssignmentBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=200, description="作业标题")
    description: Optional[str] = Field(None, description="作业描述")
    content_id: Optional[int] = Field(None, description="关联内容ID")
    deadline: datetime = Field(..., description="截止时间")


class AssignmentCreate(AssignmentBase):
    pass


class AssignmentUpdate(BaseModel):
    title: Optional[str] = Field(
        None, min_length=1, max_length=200, description="作业标题"
    )
    description: Optional[str] = Field(None, description="作业描述")
    content_id: Optional[int] = Field(None, description="关联内容ID")
    deadline: Optional[datetime] = Field(None, description="截止时间")
    status: Optional[str] = Field(
        None, pattern="^(active|completed|archived)$", description="状态"
    )


class AssignmentResponse(AssignmentBase):
    id: int
    status: str
    created_by: str
    created_at: datetime
    updated_at: Optional[datetime] = None
    submission_count: Optional[int] = 0

    class Config:
        from_attributes = True


class SubmissionBase(BaseModel):
    assignment_id: int = Field(..., description="作业 ID")
    content: Optional[str] = Field(None, description="提交内容（如果是文本）")


class SubmissionCreate(SubmissionBase):
    file_path: Optional[str] = Field(None, description="文件存储路径")
    file_name: Optional[str] = Field(None, description="文件名称")
    file_size: Optional[int] = Field(None, description="文件大小 (bytes)")


class SubmissionUpdate(BaseModel):
    status: Optional[str] = Field(
        None, pattern="^(submitted|reviewed|completed)$", description="状态"
    )
    grade: Optional[int] = Field(None, ge=0, le=100, description="评分")
    feedback: Optional[str] = Field(None, description="反馈")
    file_path: Optional[str] = Field(None, description="文件存储路径")
    file_name: Optional[str] = Field(None, description="文件名称")
    file_size: Optional[int] = Field(None, description="文件大小 (bytes)")
    content: Optional[str] = Field(None, description="提交内容（如果是文本）")


class SubmissionResponse(BaseModel):
    id: int
    assignment_id: int
    user_id: str
    file_path: Optional[str] = None
    file_name: Optional[str] = None
    file_size: Optional[int] = None
    file_url: Optional[str] = None
    content: Optional[str] = None
    status: str
    grade: Optional[int] = None
    feedback: Optional[str] = None
    submitted_at: datetime
    reviewed_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    username: Optional[str] = None

    class Config:
        from_attributes = True


class AssignmentWithSubmissionsResponse(AssignmentResponse):
    submissions: List[SubmissionResponse] = []


class FileUploadResponse(BaseModel):
    file_path: str
    file_name: str
    file_size: int
    file_url: Optional[str] = None
