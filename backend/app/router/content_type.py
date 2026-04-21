from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import Optional
from app.database import get_db
from app.schema.content_type import (
    ContentTypeCreate, ContentTypeUpdate, ContentTypeResponse, ContentTypeListResponse
)
from app.service.content_type import ContentTypeService

router = APIRouter()
type_service = ContentTypeService()


@router.post("/", response_model=ContentTypeResponse, status_code=status.HTTP_201_CREATED)
def create_content_type(content_type: ContentTypeCreate, db: Session = Depends(get_db)):
    if type_service.get_by_code(db, content_type.code):
        raise HTTPException(status_code=400, detail="Type code already exists")
    
    db_type = type_service.create_type(db, content_type)
    return db_type


@router.get("/", response_model=ContentTypeListResponse)
def get_content_types(
    skip: int = 0,
    limit: int = 100,
    is_active: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    total = type_service.get_types_count(db, is_active)
    types = type_service.get_all_types(db, skip, limit, is_active)
    return {"total": total, "types": types}


@router.get("/{type_id}", response_model=ContentTypeResponse)
def get_content_type(type_id: str, db: Session = Depends(get_db)):
    db_type = type_service.get_by_id(db, type_id)
    if not db_type:
        raise HTTPException(status_code=404, detail="Content type not found")
    return db_type


@router.put("/{type_id}", response_model=ContentTypeResponse)
def update_content_type(type_id: str, content_type: ContentTypeUpdate, db: Session = Depends(get_db)):
    db_type = type_service.update_type(db, type_id, content_type)
    if not db_type:
        raise HTTPException(status_code=404, detail="Content type not found")
    return db_type


@router.delete("/{type_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_content_type(type_id: str, db: Session = Depends(get_db)):
    success = type_service.delete_type(db, type_id)
    if not success:
        raise HTTPException(status_code=404, detail="Content type not found")
