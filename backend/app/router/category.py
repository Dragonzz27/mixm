from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models.category import Category
from app.schema.category import (
    CategoryCreate,
    CategoryUpdate,
    CategoryResponse,
    CategoryTreeResponse,
)
from app.service.category import CategoryService

router = APIRouter()
category_service = CategoryService()


@router.post("/", response_model=CategoryResponse, status_code=status.HTTP_201_CREATED)
def create_category(category: CategoryCreate, db: Session = Depends(get_db)):
    if category.parent_id:
        parent = category_service.get_category_by_id(db, category.parent_id)
        if not parent:
            raise HTTPException(status_code=404, detail="Parent category not found")

    db_category = category_service.create_category(db, category)
    return db_category


@router.get("/", response_model=List[CategoryResponse])
def get_categories(
    parent_id: int = None,
    skip: int = 0,
    limit: int = 100,
    keyword: str = None,
    top_level: bool = False,
    db: Session = Depends(get_db),
):
    if top_level:
        categories = category_service.get_categories_by_parent(db, None)
    elif parent_id is not None:
        categories = category_service.get_categories_by_parent(db, parent_id)
    else:
        categories = category_service.get_all_categories(
            db, skip=skip, limit=limit, keyword=keyword
        )
    return categories


@router.get("/tree", response_model=List[CategoryTreeResponse])
def get_category_tree(
    parent_id: int = None, max_depth: int = 3, db: Session = Depends(get_db)
):
    if parent_id is not None:
        parent = category_service.get_category_by_id(db, parent_id)
        if not parent:
            raise HTTPException(status_code=404, detail="Parent category not found")

    return category_service.get_category_tree(db, parent_id, max_depth)


@router.get("/{category_id}", response_model=CategoryResponse)
def get_category(category_id: int, db: Session = Depends(get_db)):
    db_category = category_service.get_category_by_id(db, category_id)
    if not db_category:
        raise HTTPException(status_code=404, detail="Category not found")
    return db_category


@router.put("/{category_id}", response_model=CategoryResponse)
def update_category(
    category_id: int, category: CategoryUpdate, db: Session = Depends(get_db)
):
    if category.parent_id:
        parent = category_service.get_category_by_id(db, category.parent_id)
        if not parent:
            raise HTTPException(status_code=404, detail="Parent category not found")

    db_category = category_service.update_category(db, category_id, category)
    if not db_category:
        raise HTTPException(status_code=404, detail="Category not found")
    return db_category


@router.delete("/{category_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_category(category_id: int, db: Session = Depends(get_db)):
    try:
        success = category_service.delete_category(db, category_id)
        if not success:
            raise HTTPException(status_code=404, detail="Category not found")
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
