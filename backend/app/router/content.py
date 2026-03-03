from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from app.database import get_db
from app.models.content import Content
from app.models.category import Category
from app.models.content_category import ContentCategory
from app.schema.content import (
    ContentCreate,
    ContentUpdate,
    ContentResponse,
    ContentListResponse,
    ContentBriefResponse,
)
from app.service.content import ContentService

router = APIRouter()
content_service = ContentService()


def get_category_tree_with_contents(
    db: Session, category_id: int, max_depth: int = 3
) -> Dict[str, Any]:
    def build_tree(cat: Category, current_depth: int = 1) -> Dict[str, Any]:
        result = {
            "id": cat.id,
            "name": cat.name,
            "description": cat.description,
            "level": cat.level,
            "depth": current_depth,
            "type": "category",
            "created_at": cat.created_at,
        }

        # Only fetch children if we haven't reached max_depth
        if current_depth < max_depth:
            children = (
                db.query(Category)
                .filter(Category.parent_id == cat.id, Category.is_active == True)
                .order_by(Category.sort_order.asc(), Category.id.asc())
                .all()
            )
            if children:
                result["children"] = [
                    build_tree(child, current_depth + 1) for child in children
                ]

        return result

    category = db.query(Category).filter(Category.id == category_id).first()
    if not category:
        return None  # type: ignore

    return build_tree(category)


@router.get("/category-tree/{category_id}")
def get_category_tree(
    category_id: int, max_depth: int = 3, db: Session = Depends(get_db)
):
    tree = get_category_tree_with_contents(db, category_id, max_depth)
    if not tree:
        raise HTTPException(status_code=404, detail="Category not found")
    return tree


@router.post("/", response_model=ContentResponse, status_code=status.HTTP_201_CREATED)
def create_content(content: ContentCreate, db: Session = Depends(get_db)):
    db_content = content_service.create_content(db, content)
    return db_content


@router.get("/", response_model=ContentListResponse)
def get_contents(
    skip: int = 0,
    limit: int = 100,
    content_type: Optional[str] = None,
    status: Optional[str] = None,
    keyword: Optional[str] = None,
    db: Session = Depends(get_db),
):
    total = content_service.get_contents_count(db, content_type, status, keyword)
    contents = content_service.get_all_contents(
        db, skip, limit, content_type, status, keyword
    )
    return {"total": total, "contents": contents}


@router.get("/by-category/{category_id}")
def get_contents_by_category(
    category_id: int,
    skip: int = 0,
    limit: int = 12,
    keyword: Optional[str] = None,
    db: Session = Depends(get_db),
):
    category = db.query(Category).filter(Category.id == category_id).first()
    if not category:
        raise HTTPException(status_code=404, detail="Category not found")

    contents = content_service.get_contents_by_category(
        db, int(category_id), limit, skip, keyword
    )
    # Use list comprehension to ensure we return ContentBriefResponse compatible data
    return {
        "category": {"id": category.id, "name": category.name},
        "contents": [ContentBriefResponse.model_validate(c) for c in contents],
    }


@router.get("/categories-with-contents")
def get_categories_with_contents(
    limit_per_category: int = 4, db: Session = Depends(get_db)
):
    categories = (
        db.query(Category)
        .filter(Category.is_active == True, Category.parent_id == None)
        .order_by(Category.sort_order.asc(), Category.id.asc())
        .all()
    )

    result = []
    for category in categories:
        contents = content_service.get_contents_by_category(
            db,
            int(category.id),
            limit_per_category,  # type: ignore
        )
        result.append(
            {
                "category": {
                    "id": category.id,
                    "name": category.name,
                    "sort_order": category.sort_order,
                },
                "contents": contents,
            }
        )

    return result


@router.get("/by-categories")
def get_contents_by_categories(
    category_ids: str = "",
    skip: int = 0,
    limit: int = 12,
    keyword: Optional[str] = None,
    db: Session = Depends(get_db),
):
    if not category_ids:
        return {"total": 0, "contents": []}

    ids = [int(x.strip()) for x in category_ids.split(",") if x.strip().isdigit()]
    if not ids:
        return {"total": 0, "contents": []}

    contents = content_service.get_contents_by_categories(db, ids, limit, skip, keyword)
    return {"total": len(contents), "contents": contents}


@router.get("/{content_id}", response_model=ContentResponse)
def get_content(content_id: int, db: Session = Depends(get_db)):
    print("here", content_id)
    db_content = content_service.get_content_by_id(db, content_id)
    if not db_content:
        raise HTTPException(status_code=404, detail="Content not found")
    return db_content


@router.put("/{content_id}", response_model=ContentResponse)
def update_content(
    content_id: int, content: ContentUpdate, db: Session = Depends(get_db)
):
    db_content = content_service.update_content(db, content_id, content)
    if not db_content:
        raise HTTPException(status_code=404, detail="Content not found")
    return db_content


@router.delete("/{content_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_content(content_id: int, db: Session = Depends(get_db)):
    success = content_service.delete_content(db, content_id)
    if not success:
        raise HTTPException(status_code=404, detail="Content not found")
