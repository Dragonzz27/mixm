import random
from typing import Dict, List, Optional, Tuple

from sqlalchemy import or_
from sqlalchemy.orm import Session, joinedload

from app.models.category import Category
from app.models.practice import Practice, PracticeQuestion, practice_roles
from app.models.question import Question
from app.models.role import Role
from app.schema.practice import PracticeCreate, PracticeUpdate
from app.service.question import question_service


class PracticeService:
    def create(self, db: Session, payload: PracticeCreate, created_by: str) -> Practice:
        practice = Practice(
            title=payload.title,
            description=payload.description,
            type_counts=payload.type_counts or {},
            created_by=created_by,
        )
        db.add(practice)
        db.commit()
        db.refresh(practice)

        self._sync_roles(db, practice, payload.visible_roles)
        self._sync_knowledge_tags(db, practice, payload.knowledge_tag_ids)
        self._sync_questions(db, practice, payload.question_ids)

        db.commit()
        db.refresh(practice)
        return self.get(db, practice.id)

    def get(self, db: Session, practice_id: str) -> Optional[Practice]:
        return (
            db.query(Practice)
            .options(
                joinedload(Practice.visible_roles),
                joinedload(Practice.knowledge_tags),
                joinedload(Practice.practice_questions).joinedload(PracticeQuestion.question).joinedload(Question.knowledge_tags),
            )
            .filter(Practice.id == practice_id)
            .first()
        )

    def get_list(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 20,
        keyword: Optional[str] = None,
        user_roles: Optional[List[str]] = None,
        is_superuser: bool = False,
    ) -> Tuple[List[Practice], int]:
        query = db.query(Practice)
        if keyword:
            query = query.filter(Practice.title.ilike(f"%{keyword}%"))

        if not is_superuser:
            if user_roles and len(user_roles) > 0:
                query = query.filter(
                    or_(
                        ~Practice.id.in_(db.query(practice_roles.c.practice_id)),
                        Practice.id.in_(
                            db.query(practice_roles.c.practice_id).filter(
                                practice_roles.c.role_id.in_(user_roles)
                            )
                        ),
                    )
                )
            else:
                query = query.filter(~Practice.id.in_(db.query(practice_roles.c.practice_id)))

        query = query.order_by(Practice.created_at.desc())
        total = query.count()
        practices = (
            query.options(
                joinedload(Practice.visible_roles),
                joinedload(Practice.knowledge_tags),
                joinedload(Practice.practice_questions).joinedload(PracticeQuestion.question).joinedload(Question.knowledge_tags),
            )
            .offset(skip)
            .limit(limit)
            .all()
        )
        return practices, total

    def update(self, db: Session, practice_id: str, payload: PracticeUpdate) -> Optional[Practice]:
        practice = self.get(db, practice_id)
        if not practice:
            return None

        update_data = payload.model_dump(exclude_unset=True)
        visible_roles = update_data.pop("visible_roles", None)
        knowledge_tag_ids = update_data.pop("knowledge_tag_ids", None)
        question_ids = update_data.pop("question_ids", None)

        for field, value in update_data.items():
            setattr(practice, field, value)

        if visible_roles is not None:
            self._sync_roles(db, practice, visible_roles)
        if knowledge_tag_ids is not None:
            self._sync_knowledge_tags(db, practice, knowledge_tag_ids)
        if question_ids is not None:
            self._sync_questions(db, practice, question_ids)

        db.commit()
        db.refresh(practice)
        return self.get(db, practice_id)

    def delete(self, db: Session, practice_id: str) -> bool:
        practice = self.get(db, practice_id)
        if not practice:
            return False
        db.delete(practice)
        db.commit()
        return True

    def auto_generate_questions(
        self,
        db: Session,
        knowledge_tag_ids: List[str],
        type_counts: Dict[str, int],
    ) -> Tuple[List[Question], Dict[str, int]]:
        shortages: Dict[str, int] = {}
        selected: List[Question] = []

        for question_type, count in (type_counts or {}).items():
            count = max(int(count or 0), 0)
            if count <= 0:
                continue

            query = (
                db.query(Question)
                .options(joinedload(Question.knowledge_tags))
                .filter(
                    Question.type == question_type,
                    Question.review_status == question_service.REVIEW_APPROVED,
                )
            )

            if knowledge_tag_ids:
                query = (
                    query.join(Question.knowledge_tags)
                    .filter(Category.id.in_(knowledge_tag_ids))
                    .distinct()
                )

            pool = query.all()
            random.shuffle(pool)
            picked = pool[:count]
            selected.extend(picked)

            if len(picked) < count:
                shortages[question_type] = count - len(picked)

        return selected, shortages

    def _sync_roles(self, db: Session, practice: Practice, role_ids: Optional[List[str]]) -> None:
        practice.visible_roles = []
        if role_ids:
            practice.visible_roles = db.query(Role).filter(Role.id.in_(role_ids)).all()

    def _sync_knowledge_tags(
        self, db: Session, practice: Practice, knowledge_tag_ids: Optional[List[str]]
    ) -> None:
        practice.knowledge_tags = []
        if knowledge_tag_ids:
            practice.knowledge_tags = (
                db.query(Category).filter(Category.id.in_(knowledge_tag_ids)).all()
            )

    def _sync_questions(self, db: Session, practice: Practice, question_ids: Optional[List[str]]) -> None:
        db.query(PracticeQuestion).filter(PracticeQuestion.practice_id == practice.id).delete()

        if not question_ids:
            return

        questions = (
            db.query(Question)
            .filter(Question.id.in_(question_ids))
            .filter(Question.review_status == question_service.REVIEW_APPROVED)
            .all()
        )
        question_map = {item.id: item for item in questions}

        for index, question_id in enumerate(question_ids):
            if question_id not in question_map:
                continue
            db.add(
                PracticeQuestion(
                    practice_id=practice.id,
                    question_id=question_id,
                    sort_order=index,
                )
            )


practice_service = PracticeService()
