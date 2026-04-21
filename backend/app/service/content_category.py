from typing import List, Optional
from sqlalchemy.orm import Session
from app.models.content_category import ContentCategory
from app.models.content import Content
from app.models.category import Category
from app.schema.content_category import ContentCategoryCreate


class ContentCategoryService:
    def get_by_id(self, db: Session, binding_id: str) -> Optional[ContentCategory]:
        return db.query(ContentCategory).filter(ContentCategory.id == binding_id).first()

    def get_by_content_id(self, db: Session, content_id: str) -> List[ContentCategory]:
        return db.query(ContentCategory).filter(ContentCategory.content_id == content_id).all()

    def get_by_category_id(self, db: Session, category_id: str) -> List[ContentCategory]:
        return db.query(ContentCategory).filter(ContentCategory.category_id == category_id).order_by(
            ContentCategory.sort_order).all()

    def get_all(self, db: Session) -> List[ContentCategory]:
        return db.query(ContentCategory).all()

    def get_binding(self, db: Session, content_id: str, category_id: str) -> Optional[ContentCategory]:
        return (db.query(ContentCategory)
                .filter(ContentCategory.content_id == content_id,
                       ContentCategory.category_id == category_id)
                .first())

    def create_binding(self, db: Session, binding: ContentCategoryCreate) -> ContentCategory:
        existing = self.get_binding(db, binding.content_id, binding.category_id)
        if existing:
            return existing
        
        db_binding = ContentCategory(
            content_id=binding.content_id,
            category_id=binding.category_id,
            sort_order=binding.sort_order
        )
        db.add(db_binding)
        db.commit()
        db.refresh(db_binding)
        return db_binding

    def delete_binding(self, db: Session, content_id: str, category_id: str) -> bool:
        binding = self.get_binding(db, content_id, category_id)
        if not binding:
            return False
        
        db.delete(binding)
        db.commit()
        return True

    def delete_all_by_content(self, db: Session, content_id: str) -> int:
        count = db.query(ContentCategory).filter(ContentCategory.content_id == content_id).count()
        db.query(ContentCategory).filter(ContentCategory.content_id == content_id).delete()
        db.commit()
        return count

    def get_content_with_categories(self, db: Session, content_id: str) -> Optional[dict]:
        content = db.query(Content).filter(Content.id == content_id).first()
        if not content:
            return None
        
        bindings = self.get_by_content_id(db, content_id)
        categories = []
        for binding in bindings:
            category = db.query(Category).filter(Category.id == binding.category_id).first()
            if category:
                categories.append({
                    "id": category.id,
                    "name": category.name,
                    "parent_id": category.parent_id,
                    "level": category.level,
                    "sort_order": binding.sort_order
                })
        
        return {
            "id": content.id,
            "title": content.title,
            "content": content.content,
            "content_type": content.content_type,
            "url": content.url,
            "status": content.status,
            "is_active": content.is_active,
            "created_at": content.created_at,
            "updated_at": content.updated_at,
            "categories": categories
        }

    def get_category_with_contents(self, db: Session, category_id: str) -> Optional[dict]:
        category = db.query(Category).filter(Category.id == category_id).first()
        if not category:
            return None
        
        bindings = self.get_by_category_id(db, category_id)
        contents = []
        for binding in bindings:
            content = db.query(Content).filter(Content.id == binding.content_id).first()
            if content:
                contents.append({
                    "id": content.id,
                    "title": content.title,
                    "content_type": content.content_type,
                    "status": content.status,
                    "sort_order": binding.sort_order
                })
        
        return {
            "id": category.id,
            "name": category.name,
            "description": category.description,
            "parent_id": category.parent_id,
            "level": category.level,
            "sort_order": category.sort_order,
            "is_active": category.is_active,
            "created_at": category.created_at,
            "updated_at": category.updated_at,
            "contents": contents
        }
