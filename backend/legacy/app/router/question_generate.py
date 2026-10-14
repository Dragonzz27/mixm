from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from pydantic import BaseModel, Field
from typing import List, Optional

from app.service.question_generate import question_generate_service

router = APIRouter()


class KnowledgeItem(BaseModel):
    name: str = Field(..., description="知识点名称")
    description: Optional[str] = Field(None, description="知识点描述")


class GenerateQuestionRequest(BaseModel):
    knowledge_items: List[KnowledgeItem] = Field(..., description="知识点列表（含名称+描述）")
    question_type: str = Field(..., description="题型")
    difficulty: int = Field(..., ge=1, le=5, description="难度1-5")


@router.post(
    "/generate",
    summary="生成题目(流式)",
)
async def generate_question(request: GenerateQuestionRequest):
    return StreamingResponse(
        question_generate_service.generate_question_stream(
            knowledge_items=[item.model_dump() for item in request.knowledge_items],
            question_type=request.question_type,
            difficulty=request.difficulty,
        ),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )