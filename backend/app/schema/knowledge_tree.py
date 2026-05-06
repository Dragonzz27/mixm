from typing import List, Optional
from pydantic import BaseModel, Field


class KnowledgeElement(BaseModel):
    name: str = Field(..., description="知识元名称")
    description: Optional[str] = Field(None, description="知识元描述/具体内容")


class KnowledgePoint(BaseModel):
    name: str = Field(..., description="知识点名称")
    description: Optional[str] = Field(None, description="知识点描述/具体内容")
    elements: List[KnowledgeElement] = Field(default_factory=list, description="知识元列表")


class Section(BaseModel):
    name: str = Field(..., description="小节名称")
    knowledge_points: List[KnowledgePoint] = Field(default_factory=list, description="知识点列表")


class Chapter(BaseModel):
    name: str = Field(..., description="章节名称")
    sections: List[Section] = Field(default_factory=list, description="小节列表")


class KnowledgeTree(BaseModel):
    course_name: str = Field(..., description="课程名称")
    total_chapters: int = Field(0, description="章节总数")
    chapters: List[Chapter] = Field(default_factory=list, description="章节列表")


class OutlineAnalysisRequest(BaseModel):
    outline_text: str = Field(..., description="考试大纲文本内容")
    course_name: Optional[str] = Field(None, description="课程名称")


class OutlineAnalysisResponse(BaseModel):
    success: bool = Field(..., description="是否成功")
    message: str = Field(..., description="返回消息")
    data: Optional[KnowledgeTree] = Field(None, description="知识树结构")


class SaveKnowledgeTreeRequest(BaseModel):
    knowledge_tree: KnowledgeTree = Field(..., description="知识树结构")


class SaveKnowledgeTreeResponse(BaseModel):
    success: bool = Field(..., description="是否成功")
    message: str = Field(..., description="返回消息")
    course_id: Optional[str] = Field(None, description="课程分类ID")