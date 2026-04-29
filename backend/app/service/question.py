from typing import List, Optional, Tuple
from sqlalchemy.orm import Session, joinedload
from sqlalchemy import and_, or_

from app.models.question import (
    Question,
    QuestionCategory,
    QuestionOption,
    QuestionAttachment,
)
from app.models.relations import question_roles, question_tags
from app.models.role import Role
from app.models.category import Category
from app.schema.question import (
    QuestionCreate,
    QuestionUpdate,
    QuestionCategoryCreate,
    QuestionCategoryUpdate,
)


class QuestionService:
    def create(
        self, db: Session, question: QuestionCreate, created_by: str
    ) -> Question:
        db_question = Question(
            title=question.title,
            type=question.type,
            content=question.content,
            answer=question.answer,
            explanation=question.explanation,
            difficulty=question.difficulty,
            tags=question.tags,
            source=question.source,
            created_by=created_by,
        )
        db.add(db_question)
        db.commit()
        db.refresh(db_question)

        if question.options:
            for opt in question.options:
                db_option = QuestionOption(
                    question_id=db_question.id,
                    option_key=opt.option_key,
                    content=opt.content,
                    is_correct=opt.is_correct,
                )
                db.add(db_option)

        if question.attachments:
            for att in question.attachments:
                db_attachment = QuestionAttachment(
                    question_id=db_question.id,
                    file_path=att.file_path,
                    file_name=att.file_name,
                    file_type=att.file_type,
                )
                db.add(db_attachment)

        if hasattr(question, 'knowledge_tag_ids') and question.knowledge_tag_ids:
            categories = db.query(Category).filter(Category.id.in_(question.knowledge_tag_ids)).all()
            db_question.knowledge_tags = categories

        if question.visible_roles and len(question.visible_roles) > 0:
            roles = db.query(Role).filter(Role.id.in_(question.visible_roles)).all()
            db_question.visible_roles = roles

        db.commit()
        db.refresh(db_question)
        return db_question

    def get(self, db: Session, question_id: str) -> Optional[Question]:
        return (
            db.query(Question)
            .options(
                joinedload(Question.options),
                joinedload(Question.attachments),
                joinedload(Question.visible_roles),
                joinedload(Question.knowledge_tags),
            )
            .filter(Question.id == question_id)
            .first()
        )

    def get_list(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 20,
        type: Optional[str] = None,
        difficulty: Optional[int] = None,
        keyword: Optional[str] = None,
        is_ai_generated: Optional[bool] = None,
        knowledge_tag_id: Optional[str] = None,
        user_roles: Optional[List[str]] = None,
        is_superuser: bool = False,
    ) -> Tuple[List[Question], int]:
        query = db.query(Question)

        if type:
            query = query.filter(Question.type == type)
        if difficulty:
            query = query.filter(Question.difficulty == difficulty)
        if keyword:
            query = query.filter(Question.title.ilike(f"%{keyword}%"))
        if is_ai_generated is not None:
            query = query.filter(Question.is_ai_generated == is_ai_generated)
        if knowledge_tag_id:
            query = query.join(Question.knowledge_tags).filter(Category.id == knowledge_tag_id)

        if not is_superuser:
            if user_roles and len(user_roles) > 0:
                query = query.filter(
                    or_(
                        ~Question.id.in_(db.query(question_roles.c.question_id)),
                        Question.id.in_(
                            db.query(question_roles.c.question_id).filter(
                                question_roles.c.role_id.in_(user_roles)
                            )
                        ),
                    )
                )
            else:
                query = query.filter(
                    ~Question.id.in_(db.query(question_roles.c.question_id))
                )

        query = query.order_by(Question.created_at.desc())
        total = query.count()
        questions = (
            query.options(
                joinedload(Question.options),
                joinedload(Question.attachments),
                joinedload(Question.visible_roles),
                joinedload(Question.knowledge_tags),
            )
            .offset(skip)
            .limit(limit)
            .all()
        )

        return questions, total

    def update(
        self, db: Session, question_id: str, question: QuestionUpdate
    ) -> Optional[Question]:
        db_question = self.get(db, question_id)
        if not db_question:
            return None

        update_data = question.model_dump(exclude_unset=True)
        visible_roles = update_data.pop("visible_roles", None)
        knowledge_tag_ids = update_data.pop("knowledge_tag_ids", None)
        options = update_data.pop("options", None)
        attachments = update_data.pop("attachments", None)

        for field, value in update_data.items():
            if value is not None:
                setattr(db_question, field, value)

        if options is not None:
            db.query(QuestionOption).filter(
                QuestionOption.question_id == question_id
            ).delete()
            for opt in options:
                db_option = QuestionOption(
                    question_id=question_id,
                    option_key=opt.option_key,
                    content=opt.content,
                    is_correct=opt.is_correct,
                )
                db.add(db_option)

        if attachments is not None:
            db.query(QuestionAttachment).filter(
                QuestionAttachment.question_id == question_id
            ).delete()
            for att in attachments:
                db_attachment = QuestionAttachment(
                    question_id=question_id,
                    file_path=att.file_path,
                    file_name=att.file_name,
                    file_type=att.file_type,
                )
                db.add(db_attachment)

        if knowledge_tag_ids is not None:
            db_question.knowledge_tags = []
            db.commit()
            if len(knowledge_tag_ids) > 0:
                categories = db.query(Category).filter(Category.id.in_(knowledge_tag_ids)).all()
                db_question.knowledge_tags = categories

        if visible_roles is not None:
            db_question.visible_roles = []
            db.commit()
            if len(visible_roles) > 0:
                roles = db.query(Role).filter(Role.id.in_(visible_roles)).all()
                db_question.visible_roles = roles

        db.commit()
        db.refresh(db_question)
        return db_question

    def delete(self, db: Session, question_id: str) -> bool:
        db_question = self.get(db, question_id)
        if not db_question:
            return False
        db.delete(db_question)
        db.commit()
        return True

    def batch_create(
        self, db: Session, questions: List[QuestionCreate], created_by: str
    ) -> List[Question]:
        created_questions = []
        for question in questions:
            db_question = self.create(db, question, created_by)
            created_questions.append(db_question)
        return created_questions


class QuestionCategoryService:
    def create(
        self, db: Session, category: QuestionCategoryCreate
    ) -> QuestionCategory:
        db_category = QuestionCategory(
            name=category.name,
            parent_id=category.parent_id,
        )
        db.add(db_category)
        db.commit()
        db.refresh(db_category)
        return db_category

    def get(self, db: Session, category_id: str) -> Optional[QuestionCategory]:
        return db.query(QuestionCategory).filter(QuestionCategory.id == category_id).first()

    def get_list(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 100,
        parent_id: Optional[str] = None,
    ) -> Tuple[List[QuestionCategory], int]:
        query = db.query(QuestionCategory)
        if parent_id:
            query = query.filter(QuestionCategory.parent_id == parent_id)
        else:
            query = query.filter(QuestionCategory.parent_id == None)

        query = query.order_by(QuestionCategory.name)
        total = query.count()
        categories = query.offset(skip).limit(limit).all()

        return categories, total

    def get_all(
        self, db: Session, parent_id: Optional[str] = None
    ) -> List[QuestionCategory]:
        query = db.query(QuestionCategory)
        if parent_id:
            query = query.filter(QuestionCategory.parent_id == parent_id)
        return query.order_by(QuestionCategory.name).all()

    def get_tree(self, db: Session) -> List[dict]:
        def build_tree(parent_id: Optional[str] = None) -> List[dict]:
            categories = self.get_all(db, parent_id)
            result = []
            for cat in categories:
                item = {
                    "id": cat.id,
                    "name": cat.name,
                    "parent_id": cat.parent_id,
                    "created_at": cat.created_at,
                    "question_count": db.query(Question)
                    .filter(Question.category_id == cat.id)
                    .count(),
                    "children": build_tree(cat.id),
                }
                result.append(item)
            return result

        return build_tree(None)

    def update(
        self, db: Session, category_id: str, category: QuestionCategoryUpdate
    ) -> Optional[QuestionCategory]:
        db_category = self.get(db, category_id)
        if not db_category:
            return None

        update_data = category.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            if value is not None:
                setattr(db_category, field, value)

        db.commit()
        db.refresh(db_category)
        return db_category

    def delete(self, db: Session, category_id: str) -> bool:
        db_category = self.get(db, category_id)
        if not db_category:
            return False

        has_children = (
            db.query(QuestionCategory)
            .filter(QuestionCategory.parent_id == category_id)
            .count()
            > 0
        )
        if has_children:
            return False

        has_questions = (
            db.query(Question).filter(Question.category_id == category_id).count() > 0
        )
        if has_questions:
            return False

        db.delete(db_category)
        db.commit()
        return True


question_service = QuestionService()
question_category_service = QuestionCategoryService()