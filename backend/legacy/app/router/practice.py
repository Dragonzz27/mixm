from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.models.user import User
from app.models.user_role import UserRole
from app.schema.practice import (
    PracticeAutoGenerateRequest,
    PracticeAutoGenerateResponse,
    PracticeCreate,
    PracticeListResponse,
    PracticeResponse,
    PracticeUpdate,
)
from app.service.auth import AuthService
from app.service.practice import practice_service

auth_service = AuthService()
router = APIRouter()
get_current_user = auth_service.get_current_user


def question_summary(question) -> dict:
    return {
        "id": question.id,
        "title": question.title,
        "type": question.type,
        "difficulty": question.difficulty,
        "review_status": question.review_status,
        "knowledge_tags": [
            {"id": tag.id, "name": tag.name}
            for tag in (question.knowledge_tags or [])
        ],
    }


def practice_to_response(practice) -> dict:
    questions = []
    for item in practice.practice_questions or []:
        if item.question:
            questions.append(question_summary(item.question))

    return {
        "id": practice.id,
        "title": practice.title,
        "description": practice.description,
        "visible_roles": [role.id for role in practice.visible_roles] if practice.visible_roles else [],
        "knowledge_tag_ids": [tag.id for tag in practice.knowledge_tags] if practice.knowledge_tags else [],
        "knowledge_tags": [
            {"id": tag.id, "name": tag.name}
            for tag in (practice.knowledge_tags or [])
        ],
        "type_counts": practice.type_counts or {},
        "created_by": practice.created_by,
        "created_at": practice.created_at,
        "updated_at": practice.updated_at,
        "question_count": len(questions),
        "questions": questions,
    }


def get_user_roles(db: Session, user_id: str) -> list[str]:
    rows = db.query(UserRole.role_id).filter(UserRole.user_id == user_id).all()
    return [row[0] for row in rows] if rows else []


@router.post("/", response_model=PracticeResponse, status_code=status.HTTP_201_CREATED)
def create_practice(
    payload: PracticeCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    practice = practice_service.create(db, payload, current_user.id)
    return practice_to_response(practice)


@router.get("/", response_model=PracticeListResponse)
def get_practices(
    skip: int = 0,
    limit: int = 20,
    keyword: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    practices, total = practice_service.get_list(
        db,
        skip=skip,
        limit=limit,
        keyword=keyword,
        user_roles=get_user_roles(db, current_user.id),
        is_superuser=current_user.is_superuser,
    )
    return {
        "total": total,
        "practices": [practice_to_response(item) for item in practices],
    }


@router.post("/auto-generate", response_model=PracticeAutoGenerateResponse)
def auto_generate_practice_questions(
    payload: PracticeAutoGenerateRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    questions, shortages = practice_service.auto_generate_questions(
        db,
        payload.knowledge_tag_ids,
        payload.type_counts,
    )
    return {
        "questions": [question_summary(item) for item in questions],
        "shortages": shortages,
    }


@router.get("/{practice_id}", response_model=PracticeResponse)
def get_practice(
    practice_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    practice = practice_service.get(db, practice_id)
    if not practice:
        raise HTTPException(status_code=404, detail="Practice not found")
    return practice_to_response(practice)


@router.put("/{practice_id}", response_model=PracticeResponse)
def update_practice(
    practice_id: str,
    payload: PracticeUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    practice = practice_service.update(db, practice_id, payload)
    if not practice:
        raise HTTPException(status_code=404, detail="Practice not found")
    return practice_to_response(practice)


@router.delete("/{practice_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_practice(
    practice_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = practice_service.delete(db, practice_id)
    if not success:
        raise HTTPException(status_code=404, detail="Practice not found")
