from typing import List, Optional
from sqlalchemy.orm import Session
from app.models.template import Template
from app.schema.template import TemplateCreate, TemplateUpdate


class TemplateService:
    def get_template_by_id(self, db: Session, template_id: int) -> Optional[Template]:
        return db.query(Template).filter(Template.id == template_id).first()

    def get_all_templates(self, db: Session, skip: int = 0, limit: int = 100) -> List[Template]:
        return db.query(Template).offset(skip).limit(limit).all()

    def search_templates(self, db: Session, keyword: str = None, skip: int = 0, limit: int = 100) -> List[Template]:
        query = db.query(Template)
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Template.name.contains(keyword))
        return query.order_by(Template.id).offset(skip).limit(limit).all()

    def get_templates_count(self, db: Session) -> int:
        return db.query(Template).count()

    def create_template(self, db: Session, template: TemplateCreate) -> Template:
        db_template = Template(
            name=template.name,
            description=template.description,
            config=template.config
        )
        db.add(db_template)
        db.commit()
        db.refresh(db_template)
        return db_template

    def update_template(self, db: Session, template_id: int, template: TemplateUpdate) -> Optional[Template]:
        db_template = self.get_template_by_id(db, template_id)
        if not db_template:
            return None
        
        update_data = template.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_template, field, value)
        
        db.commit()
        db.refresh(db_template)
        return db_template

    def delete_template(self, db: Session, template_id: int) -> bool:
        db_template = self.get_template_by_id(db, template_id)
        if not db_template:
            return False
        
        db.delete(db_template)
        db.commit()
        return True
