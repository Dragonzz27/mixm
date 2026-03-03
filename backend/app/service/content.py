from typing import List, Optional
from sqlalchemy.orm import Session, joinedload
from app.models.content import Content
from app.models.content_category import ContentCategory
from app.schema.content import ContentCreate, ContentUpdate


class ContentService:
    def get_content_by_id(self, db: Session, content_id: int) -> Optional[Content]:
        print(db.query(Content).filter(Content.id == content_id))
        return db.query(Content).filter(Content.id == content_id).first()

    def get_all_contents(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 100,
        content_type: Optional[str] = None,
        status: Optional[str] = None,
        keyword: Optional[str] = None,
    ) -> List[Content]:
        query = db.query(Content).filter(Content.is_active == True)
        if content_type:
            query = query.filter(Content.content_type == content_type)
        if status:
            query = query.filter(Content.status == status)
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Content.title.contains(keyword))
        return query.order_by(Content.created_at.desc()).offset(skip).limit(limit).all()

    def get_contents_by_category(
        self,
        db: Session,
        category_id: int,
        limit: int = 12,
        skip: int = 0,
        keyword: Optional[str] = None,
    ) -> List[Content]:
        query = (
            db.query(Content)
            .join(ContentCategory, Content.id == ContentCategory.content_id)
            .filter(
                ContentCategory.category_id == category_id,
                Content.is_active == True,
                Content.status == "published",
            )
        )
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Content.title.contains(keyword))

        return (
            query.order_by(ContentCategory.sort_order.asc(), Content.created_at.desc())
            .offset(skip)
            .limit(limit)
            .all()
        )

    def get_contents_by_categories(
        self,
        db: Session,
        category_ids: List[int],
        limit: int = 12,
        skip: int = 0,
        keyword: Optional[str] = None,
    ) -> List[Content]:
        query = (
            db.query(Content)
            .join(ContentCategory, Content.id == ContentCategory.content_id)
            .filter(
                ContentCategory.category_id.in_(category_ids),
                Content.is_active == True,
                Content.status == "published",
            )
        )
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Content.title.contains(keyword))

        return (
            query.order_by(ContentCategory.sort_order.asc(), Content.created_at.desc())
            .offset(skip)
            .limit(limit)
            .all()
        )

    def get_contents_count(
        self,
        db: Session,
        content_type: Optional[str] = None,
        status: Optional[str] = None,
        keyword: Optional[str] = None,
    ) -> int:
        query = db.query(Content).filter(Content.is_active == True)
        if content_type:
            query = query.filter(Content.content_type == content_type)
        if status:
            query = query.filter(Content.status == status)
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Content.title.contains(keyword))
        return query.count()

    def create_content(self, db: Session, content: ContentCreate) -> Content:
        db_content = Content(
            title=content.title,
            content=content.content,
            content_type=content.content_type,
            url=content.url,
            status=content.status,
            template_id=content.template_id,
            is_active=content.is_active,
        )
        db.add(db_content)
        db.commit()
        db.refresh(db_content)
        return db_content

    def update_content(
        self, db: Session, content_id: int, content: ContentUpdate
    ) -> Optional[Content]:
        db_content = self.get_content_by_id(db, content_id)
        if not db_content:
            return None

        update_data = content.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_content, field, value)

        db.commit()
        db.refresh(db_content)
        return db_content

    def delete_content(self, db: Session, content_id: int) -> bool:
        db_content = self.get_content_by_id(db, content_id)
        if not db_content:
            return False

        db.delete(db_content)
        db.commit()
        return True
