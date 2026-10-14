from fastapi import APIRouter, Depends, status
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session

from app.database import get_db
from app.schema.knowledge_tree import (
    OutlineAnalysisRequest,
    OutlineAnalysisResponse,
    SaveKnowledgeTreeRequest,
    SaveKnowledgeTreeResponse,
)
from app.service.knowledge_tree import knowledge_tree_service

router = APIRouter()


@router.post(
    "/analyze",
    summary="分析考试大纲生成知识树(流式)",
)
async def analyze_outline_stream(request: OutlineAnalysisRequest):
    return StreamingResponse(
        knowledge_tree_service.analyze_outline_stream(request),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


@router.post(
    "/save",
    response_model=SaveKnowledgeTreeResponse,
    summary="保存知识树到数据库",
)
async def save_knowledge_tree(
    request: SaveKnowledgeTreeRequest,
    db: Session = Depends(get_db),
) -> SaveKnowledgeTreeResponse:
    try:
        course_id = knowledge_tree_service.save_to_database(
            db, request.knowledge_tree
        )
        return SaveKnowledgeTreeResponse(
            success=True,
            message="知识树保存成功",
            course_id=course_id,
        )
    except Exception as e:
        return SaveKnowledgeTreeResponse(
            success=False,
            message=f"保存失败: {str(e)}",
            course_id=None,
        )