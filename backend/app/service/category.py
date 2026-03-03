from typing import List, Optional
from sqlalchemy.orm import Session
from app.models.category import Category
from app.schema.category import CategoryCreate, CategoryUpdate


class CategoryService:
    def get_category_by_id(self, db: Session, category_id: int) -> Optional[Category]:
        return db.query(Category).filter(Category.id == category_id).first()

    def get_categories_by_parent(
        self, db: Session, parent_id: Optional[int] = None
    ) -> List[Category]:
        query = db.query(Category).filter(
            Category.parent_id == parent_id, Category.is_active == True
        )
        return query.order_by(Category.sort_order, Category.id).all()

    def get_all_categories(
        self, db: Session, skip: int = 0, limit: int = 100, keyword: str = None
    ) -> List[Category]:
        query = db.query(Category).filter(Category.is_active == True)
        if keyword:
            keyword = keyword.strip()
            if keyword:
                query = query.filter(Category.name.contains(keyword))
        return (
            query.order_by(Category.sort_order, Category.id)
            .offset(skip)
            .limit(limit)
            .all()
        )

    def get_category_tree(
        self,
        db: Session,
        parent_id: Optional[int] = None,
        max_depth: int = 3,
        current_depth: int = 1,
    ) -> List[dict]:
        categories = self.get_categories_by_parent(db, parent_id)
        result = []
        for category in categories:
            category_dict = {
                "id": category.id,
                "name": category.name,
                "description": category.description,
                "parent_id": category.parent_id,
                "level": category.level,
                "sort_order": category.sort_order,
                "is_active": category.is_active,
                "created_at": category.created_at,
                "updated_at": category.updated_at,
            }

            if current_depth < max_depth:
                children = self.get_category_tree(
                    db, category.id, max_depth, current_depth + 1
                )
                category_dict["children"] = children
            else:
                category_dict["children"] = []

            result.append(category_dict)
        return result

    def create_category(self, db: Session, category: CategoryCreate) -> Category:
        level = 1
        if category.parent_id:
            parent = self.get_category_by_id(db, category.parent_id)
            if parent:
                level = parent.level + 1

        db_category = Category(
            name=category.name,
            description=category.description,
            parent_id=category.parent_id,
            level=level,
            sort_order=category.sort_order,
            is_active=category.is_active,
        )
        db.add(db_category)
        db.commit()
        db.refresh(db_category)
        return db_category

    def update_category(
        self, db: Session, category_id: int, category: CategoryUpdate
    ) -> Optional[Category]:
        db_category = self.get_category_by_id(db, category_id)
        if not db_category:
            return None

        level = db_category.level
        if category.parent_id and category.parent_id != db_category.parent_id:
            parent = self.get_category_by_id(db, category.parent_id)
            if parent:
                level = parent.level + 1

        update_data = category.model_dump(exclude_unset=True)
        update_data["level"] = level

        for field, value in update_data.items():
            setattr(db_category, field, value)

        db.commit()
        db.refresh(db_category)
        return db_category

    def delete_category(self, db: Session, category_id: int) -> bool:
        db_category = self.get_category_by_id(db, category_id)
        if not db_category:
            return False

        children = self.get_categories_by_parent(db, category_id)
        if children:
            raise ValueError("Cannot delete category with children")

        db.delete(db_category)
        db.commit()
        return True

    def update_children_level(self, db: Session, category_id: int, level: int):
        children = self.get_categories_by_parent(db, category_id)
        for child in children:
            child.level = level + 1
            db.commit()
            self.update_children_level(db, child.id, child.level)
