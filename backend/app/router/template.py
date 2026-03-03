from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models.template import Template
from app.schema.template import TemplateCreate, TemplateUpdate, TemplateResponse, TemplateListResponse
from app.service.template import TemplateService

router = APIRouter()
template_service = TemplateService()


@router.post("/", response_model=TemplateResponse, status_code=status.HTTP_201_CREATED)
def create_template(template: TemplateCreate, db: Session = Depends(get_db)):
    db_template = db.query(Template).filter(Template.name == template.name).first()
    if db_template:
        raise HTTPException(status_code=400, detail="Template name already exists")
    
    db_template = template_service.create_template(db, template)
    return db_template


@router.get("/", response_model=TemplateListResponse)
def get_templates(skip: int = 0, limit: int = 100, keyword: str = None, db: Session = Depends(get_db)):
    if keyword:
        total = len(template_service.search_templates(db, keyword))
        templates = template_service.search_templates(db, keyword, skip, limit)
    else:
        total = template_service.get_templates_count(db)
        templates = template_service.get_all_templates(db, skip, limit)
    return {"total": total, "templates": templates}


@router.get("/{template_id}", response_model=TemplateResponse)
def get_template(template_id: int, db: Session = Depends(get_db)):
    db_template = template_service.get_template_by_id(db, template_id)
    if not db_template:
        raise HTTPException(status_code=404, detail="Template not found")
    return db_template


@router.put("/{template_id}", response_model=TemplateResponse)
def update_template(template_id: int, template: TemplateUpdate, db: Session = Depends(get_db)):
    db_template = template_service.update_template(db, template_id, template)
    if not db_template:
        raise HTTPException(status_code=404, detail="Template not found")
    return db_template


@router.delete("/{template_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_template(template_id: int, db: Session = Depends(get_db)):
    success = template_service.delete_template(db, template_id)
    if not success:
        raise HTTPException(status_code=404, detail="Template not found")
