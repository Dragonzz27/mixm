from typing import List, Optional
from sqlalchemy.orm import Session
from app.models.content_type import ContentType
from app.schema.content_type import ContentTypeCreate, ContentTypeUpdate


class ContentTypeService:
    def get_by_id(self, db: Session, type_id: str) -> Optional[ContentType]:
        return db.query(ContentType).filter(ContentType.id == type_id).first()

    def get_by_code(self, db: Session, code: str) -> Optional[ContentType]:
        return db.query(ContentType).filter(ContentType.code == code).first()

    def get_all_types(self, db: Session, skip: int = 0, limit: int = 100, 
                      is_active: Optional[bool] = None) -> List[ContentType]:
        query = db.query(ContentType)
        if is_active is not None:
            query = query.filter(ContentType.is_active == is_active)
        return query.offset(skip).limit(limit).all()

    def get_types_count(self, db: Session, is_active: Optional[bool] = None) -> int:
        query = db.query(ContentType)
        if is_active is not None:
            query = query.filter(ContentType.is_active == is_active)
        return query.count()

    def create_type(self, db: Session, content_type: ContentTypeCreate) -> ContentType:
        db_type = ContentType(
            name=content_type.name,
            code=content_type.code,
            description=content_type.description,
            is_active=content_type.is_active
        )
        db.add(db_type)
        db.commit()
        db.refresh(db_type)
        return db_type

    def update_type(self, db: Session, type_id: str, content_type: ContentTypeUpdate) -> Optional[ContentType]:
        db_type = self.get_by_id(db, type_id)
        if not db_type:
            return None
        
        update_data = content_type.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_type, field, value)
        
        db.commit()
        db.refresh(db_type)
        return db_type

    def delete_type(self, db: Session, type_id: str) -> bool:
        db_type = self.get_by_id(db, type_id)
        if not db_type:
            return False
        
        db.delete(db_type)
        db.commit()
        return True
