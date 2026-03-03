from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import Optional
from app.database import get_db
from app.schema.content_category import (
    ContentCategoryCreate, ContentCategoryResponse, 
    ContentWithCategories, CategoryWithContents
)
from app.service.content_category import ContentCategoryService
from app.service.content import ContentService
from app.models.content_category import ContentCategory

router = APIRouter()
binding_service = ContentCategoryService()
content_service = ContentService()


@router.post("/", response_model=ContentCategoryResponse, status_code=status.HTTP_201_CREATED)
def create_binding(binding: ContentCategoryCreate, db: Session = Depends(get_db)):
    if not content_service.get_content_by_id(db, binding.content_id):
        raise HTTPException(status_code=404, detail="Content not found")
    
    from app.models.category import Category
    if not db.query(Category).filter(Category.id == binding.category_id).first():
        raise HTTPException(status_code=404, detail="Category not found")
    
    db_binding = binding_service.create_binding(db, binding)
    return db_binding


@router.get("/{binding_id}", response_model=ContentCategoryResponse)
def get_binding(binding_id: int, db: Session = Depends(get_db)):
    db_binding = binding_service.get_by_id(db, binding_id)
    if not db_binding:
        raise HTTPException(status_code=404, detail="Binding not found")
    return db_binding


@router.get("/content/{content_id}", response_model=list[ContentCategoryResponse])
def get_bindings_by_content(content_id: int, db: Session = Depends(get_db)):
    if not content_service.get_content_by_id(db, content_id):
        raise HTTPException(status_code=404, detail="Content not found")
    
    return binding_service.get_by_content_id(db, content_id)


@router.get("/", response_model=list[ContentCategoryResponse])
def get_all_bindings(db: Session = Depends(get_db)):
    return binding_service.get_all(db)


@router.get("/category/{category_id}", response_model=list[ContentCategoryResponse])
def get_bindings_by_category(category_id: int, db: Session = Depends(get_db)):
    from app.models.category import Category
    if not db.query(Category).filter(Category.id == category_id).first():
        raise HTTPException(status_code=404, detail="Category not found")
    
    return binding_service.get_by_category_id(db, category_id)


@router.delete("/{binding_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_binding(binding_id: int, db: Session = Depends(get_db)):
    db_binding = binding_service.get_by_id(db, binding_id)
    if not db_binding:
        raise HTTPException(status_code=404, detail="Binding not found")
    
    success = binding_service.delete_binding(db, int(db_binding.content_id), int(db_binding.category_id))
    if not success:
        raise HTTPException(status_code=404, detail="Binding not found")


@router.delete("/by-content/{content_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_all_bindings_by_content(content_id: int, db: Session = Depends(get_db)):
    if not content_service.get_content_by_id(db, content_id):
        raise HTTPException(status_code=404, detail="Content not found")
    
    binding_service.delete_all_by_content(db, content_id)


@router.get("/detail/content/{content_id}", response_model=ContentWithCategories)
def get_content_with_categories(content_id: int, db: Session = Depends(get_db)):
    result = binding_service.get_content_with_categories(db, content_id)
    if not result:
        raise HTTPException(status_code=404, detail="Content not found")
    return result


@router.get("/detail/category/{category_id}", response_model=CategoryWithContents)
def get_category_with_contents(category_id: int, db: Session = Depends(get_db)):
    result = binding_service.get_category_with_contents(db, category_id)
    if not result:
        raise HTTPException(status_code=404, detail="Category not found")
    return result


@router.get("/category/{category_id}/unbound")
def get_unbound_contents(category_id: int, keyword: str = None, skip: int = 0, limit: int = 20, db: Session = Depends(get_db)):
    from app.models.category import Category
    from app.models.content import Content
    
    if not db.query(Category).filter(Category.id == category_id).first():
        raise HTTPException(status_code=404, detail="Category not found")
    
    bound_content_ids = set(b.id for b in binding_service.get_by_category_id(db, category_id))
    
    query = db.query(Content).filter(~Content.id.in_(bound_content_ids))
    
    if keyword:
        keyword = keyword.strip()
        if keyword:
            query = query.filter(Content.title.contains(keyword))
    
    total = query.count()
    contents = query.order_by(Content.id).offset(skip).limit(limit).all()
    
    return {
        "total": total,
        "contents": [{"id": c.id, "title": c.title, "content_type": c.content_type, "status": c.status} for c in contents]
    }


@router.post("/category/{category_id}/batch-bind")
def batch_bind_contents(category_id: int, data: dict, db: Session = Depends(get_db)):
    from app.models.category import Category
    
    if not db.query(Category).filter(Category.id == category_id).first():
        raise HTTPException(status_code=404, detail="Category not found")
    
    content_ids = data.get("content_ids", [])
    for content_id in content_ids:
        binding = binding_service.get_binding(db, content_id, category_id)
        if not binding:
            db_binding = ContentCategory(
                content_id=content_id,
                category_id=category_id
            )
            db.add(db_binding)
    
    db.commit()
    return {"message": "Batch bind successful", "count": len(content_ids)}


@router.delete("/category/{category_id}/batch-unbind")
def batch_unbind_contents(category_id: int, data: dict, db: Session = Depends(get_db)):
    content_ids = data.get("content_ids", [])
    for content_id in content_ids:
        binding_service.delete_binding(db, content_id, category_id)
    
    return {"message": "Batch unbind successful", "count": len(content_ids)}


@router.post("/counts")
def get_category_binding_counts(data: dict, db: Session = Depends(get_db)):
    category_ids = data.get("category_ids", [])
    if not category_ids:
        return {}
    
    result = {}
    for category_id in category_ids:
        count = db.query(ContentCategory).filter(ContentCategory.category_id == int(category_id)).count()
        result[int(category_id)] = count
    
    return result
