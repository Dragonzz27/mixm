from datetime import datetime
from typing import Dict, List, Optional

from pydantic import BaseModel, Field


class PracticeQuestionSummary(BaseModel):
    id: str
    title: str
    type: str
    difficulty: int
    review_status: str
    knowledge_tags: List[dict] = []


class PracticeBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = None
    visible_roles: Optional[List[str]] = None
    knowledge_tag_ids: Optional[List[str]] = None
    type_counts: Optional[Dict[str, int]] = None


class PracticeCreate(PracticeBase):
    question_ids: List[str] = Field(default_factory=list)


class PracticeUpdate(BaseModel):
    title: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = None
    visible_roles: Optional[List[str]] = None
    knowledge_tag_ids: Optional[List[str]] = None
    type_counts: Optional[Dict[str, int]] = None
    question_ids: Optional[List[str]] = None


class PracticeResponse(PracticeBase):
    id: str
    created_by: str
    created_at: datetime
    updated_at: Optional[datetime] = None
    question_count: int = 0
    questions: List[PracticeQuestionSummary] = []
    knowledge_tags: List[dict] = []

    class Config:
        from_attributes = True


class PracticeListResponse(BaseModel):
    total: int
    practices: List[PracticeResponse]


class PracticeAutoGenerateRequest(BaseModel):
    knowledge_tag_ids: List[str] = Field(default_factory=list)
    type_counts: Dict[str, int] = Field(default_factory=dict)


class PracticeAutoGenerateResponse(BaseModel):
    questions: List[PracticeQuestionSummary]
    shortages: Dict[str, int]
