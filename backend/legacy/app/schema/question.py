from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field


class QuestionOptionBase(BaseModel):
    option_key: str = Field(..., description="选项标识 A/B/C/D")
    content: str = Field(..., description="选项内容")
    is_correct: bool = Field(default=False, description="是否正确答案")


class QuestionOptionCreate(QuestionOptionBase):
    pass


class QuestionOptionResponse(QuestionOptionBase):
    id: str

    class Config:
        from_attributes = True


class QuestionAttachmentBase(BaseModel):
    file_path: Optional[str] = Field(None, description="文件存储路径")
    file_name: Optional[str] = Field(None, description="文件名称")
    file_type: Optional[str] = Field(None, description="文件类型")


class QuestionAttachmentCreate(QuestionAttachmentBase):
    pass


class QuestionAttachmentResponse(QuestionAttachmentBase):
    id: str

    class Config:
        from_attributes = True


class QuestionBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=500, description="题目内容/题干")
    type: str = Field(
        ...,
        description="题型: single_choice/multi_choice/true_false/blank/essay"
    )
    content: Optional[str] = Field(None, description="详细解析/内容")
    answer: Optional[str] = Field(None, description="答案")
    explanation: Optional[str] = Field(None, description="解析")
    difficulty: int = Field(..., ge=1, le=5, description="难度 1-5")
    tags: Optional[List[str]] = Field(None, description="简单标签列表 (非知识元)")
    source: Optional[str] = Field(None, max_length=100, description="来源")
    knowledge_tag_ids: Optional[List[str]] = Field(
        None, description="知识元/知识点分类ID列表"
    )
    visible_roles: Optional[List[str]] = Field(
        None, description="可见角色 ID 列表 (为空表示所有角色可见)"
    )


class QuestionCreate(QuestionBase):
    options: Optional[List[QuestionOptionCreate]] = Field(
        None, description="选择题选项"
    )
    attachments: Optional[List[QuestionAttachmentCreate]] = Field(
        None, description="附件列表"
    )


class QuestionUpdate(BaseModel):
    title: Optional[str] = Field(None, min_length=1, max_length=500, description="题目内容/题干")
    type: Optional[str] = Field(None, description="题型")
    content: Optional[str] = Field(None, description="详细解析/内容")
    answer: Optional[str] = Field(None, description="答案")
    explanation: Optional[str] = Field(None, description="解析")
    difficulty: Optional[int] = Field(None, ge=1, le=5, description="难度 1-5")
    tags: Optional[List[str]] = Field(None, description="简单标签列表")
    source: Optional[str] = Field(None, max_length=100, description="来源")
    knowledge_tag_ids: Optional[List[str]] = Field(None, description="知识元/知识点分类ID列表")
    visible_roles: Optional[List[str]] = Field(None, description="可见角色 ID 列表")
    options: Optional[List[QuestionOptionCreate]] = Field(None, description="选择题选项")
    attachments: Optional[List[QuestionAttachmentCreate]] = Field(None, description="附件列表")


class QuestionReviewRequest(BaseModel):
    review_status: str = Field(..., description="pending/approved/rejected")


class KnowledgeTagResponse(BaseModel):
    id: str
    name: str

    class Config:
        from_attributes = True


class QuestionResponse(QuestionBase):
    id: str
    is_ai_generated: bool
    review_status: str
    reviewed_by: Optional[str] = None
    reviewed_at: Optional[datetime] = None
    created_by: Optional[str] = None
    created_at: datetime
    updated_at: Optional[datetime] = None
    options: List[QuestionOptionResponse] = []
    attachments: List[QuestionAttachmentResponse] = []
    knowledge_tags: List[KnowledgeTagResponse] = []

    class Config:
        from_attributes = True


class QuestionListResponse(BaseModel):
    total: int
    questions: List[QuestionResponse]


class QuestionCategoryBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100, description="分类名称")
    parent_id: Optional[str] = Field(None, description="父分类ID")


class QuestionCategoryCreate(QuestionCategoryBase):
    pass


class QuestionCategoryUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100, description="分类名称")
    parent_id: Optional[str] = Field(None, description="父分类ID")


class QuestionCategoryResponse(QuestionCategoryBase):
    id: str
    created_at: datetime

    class Config:
        from_attributes = True


class QuestionCategoryTreeResponse(QuestionCategoryResponse):
    children: List["QuestionCategoryTreeResponse"] = []
    question_count: int = 0

    class Config:
        from_attributes = True


class QuestionCategoryListResponse(BaseModel):
    total: int
    categories: List[QuestionCategoryResponse]


class QuestionGenerateRequest(BaseModel):
    topic: str = Field(..., description="主题/知识点")
    difficulty: int = Field(..., ge=1, le=5, description="目标难度 1-5")
    count: int = Field(default=10, ge=1, le=50, description="生成数量")
    types: List[str] = Field(
        default=["single_choice"], description="题型要求"
    )
    category_id: Optional[str] = Field(None, description="分类ID")
