from typing import Optional, List

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.models.user import User
from app.schema.question import (
    QuestionCreate,
    QuestionUpdate,
    QuestionResponse,
    QuestionListResponse,
    QuestionCategoryCreate,
    QuestionCategoryUpdate,
    QuestionCategoryResponse,
    QuestionCategoryListResponse,
    QuestionCategoryTreeResponse,
)
from app.service.question import question_service, question_category_service
from app.service.auth import AuthService

auth_service = AuthService()
router = APIRouter()

get_current_user = auth_service.get_current_user


def question_to_response(question) -> dict:
    return {
        "id": question.id,
        "title": question.title,
        "type": question.type,
        "content": question.content,
        "answer": question.answer,
        "explanation": question.explanation,
        "difficulty": question.difficulty,
        "tags": question.tags or [],
        "source": question.source,
        "is_ai_generated": question.is_ai_generated,
        "visible_roles": [role.id for role in question.visible_roles]
        if question.visible_roles
        else [],
        "knowledge_tags": [
            {"id": tag.id, "name": tag.name}
            for tag in question.knowledge_tags
        ]
        if question.knowledge_tags
        else [],
        "created_by": question.created_by,
        "created_at": question.created_at,
        "updated_at": question.updated_at,
        "options": [
            {
                "id": opt.id,
                "option_key": opt.option_key,
                "content": opt.content,
                "is_correct": opt.is_correct,
            }
            for opt in question.options
        ]
        if question.options
        else [],
        "attachments": [
            {
                "id": att.id,
                "file_path": att.file_path,
                "file_name": att.file_name,
                "file_type": att.file_type,
            }
            for att in question.attachments
        ]
        if question.attachments
        else [],
    }


@router.post(
    "/", response_model=QuestionResponse, status_code=status.HTTP_201_CREATED
)
def create_question(
    question: QuestionCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_question = question_service.create(db, question, current_user.id)
    return question_to_response(db_question)


@router.get("/", response_model=QuestionListResponse)
def get_questions(
    skip: int = 0,
    limit: int = 20,
    type: Optional[str] = None,
    difficulty: Optional[int] = None,
    keyword: Optional[str] = None,
    is_ai_generated: Optional[bool] = None,
    knowledge_tag_id: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    from app.models.user_role import UserRole

    user_roles = (
        db.query(UserRole.role_id).filter(UserRole.user_id == current_user.id).all()
    )
    role_list = [r[0] for r in user_roles] if user_roles else []

    questions, total = question_service.get_list(
        db,
        skip,
        limit,
        type=type,
        difficulty=difficulty,
        keyword=keyword,
        is_ai_generated=is_ai_generated,
        knowledge_tag_id=knowledge_tag_id,
        user_roles=role_list,
        is_superuser=current_user.is_superuser,
    )
    return {
        "total": total,
        "questions": [question_to_response(q) for q in questions],
    }


@router.get("/ai-generated", response_model=QuestionListResponse)
def get_ai_generated_questions(
    skip: int = 0,
    limit: int = 20,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    from app.models.user_role import UserRole

    user_roles = (
        db.query(UserRole.role_id).filter(UserRole.user_id == current_user.id).all()
    )
    role_list = [r[0] for r in user_roles] if user_roles else []

    questions, total = question_service.get_list(
        db,
        skip,
        limit,
        is_ai_generated=True,
        user_roles=role_list,
        is_superuser=current_user.is_superuser,
    )
    return {
        "total": total,
        "questions": [question_to_response(q) for q in questions],
    }


@router.get("/{question_id}", response_model=QuestionResponse)
def get_question(
    question_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    question = question_service.get(db, question_id)
    if not question:
        raise HTTPException(status_code=404, detail="Question not found")
    return question_to_response(question)


@router.put("/{question_id}", response_model=QuestionResponse)
def update_question(
    question_id: str,
    question: QuestionUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_question = question_service.update(db, question_id, question)
    if not db_question:
        raise HTTPException(status_code=404, detail="Question not found")
    return question_to_response(db_question)


@router.delete("/{question_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_question(
    question_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = question_service.delete(db, question_id)
    if not success:
        raise HTTPException(status_code=404, detail="Question not found")


@router.post("/batch", response_model=list[QuestionResponse])
def batch_create_questions(
    questions: List[QuestionCreate],
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    created_questions = question_service.batch_create(
        db, questions, current_user.id
    )
    return [question_to_response(q) for q in created_questions]


category_router = APIRouter()


@category_router.post(
    "/", response_model=QuestionCategoryResponse, status_code=status.HTTP_201_CREATED
)
def create_category(
    category: QuestionCategoryCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_category = question_category_service.create(db, category)
    return db_category


@category_router.get("/", response_model=QuestionCategoryListResponse)
def get_categories(
    skip: int = 0,
    limit: int = 100,
    parent_id: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    categories, total = question_category_service.get_list(
        db, skip, limit, parent_id
    )
    return {"total": total, "categories": categories}


@category_router.get("/tree")
def get_category_tree(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    return question_category_service.get_tree(db)


@category_router.get("/{category_id}", response_model=QuestionCategoryResponse)
def get_category(
    category_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    category = question_category_service.get(db, category_id)
    if not category:
        raise HTTPException(status_code=404, detail="Category not found")
    return category


@category_router.put("/{category_id}", response_model=QuestionCategoryResponse)
def update_category(
    category_id: str,
    category: QuestionCategoryUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_category = question_category_service.update(db, category_id, category)
    if not db_category:
        raise HTTPException(status_code=404, detail="Category not found")
    return db_category


@category_router.delete("/{category_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_category(
    category_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = question_category_service.delete(db, category_id)
    if not success:
        raise HTTPException(
            status_code=400,
            detail="Cannot delete category (has children or questions)",
        )