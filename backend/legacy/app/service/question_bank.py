import random
from typing import Dict, List, Optional, Tuple

from sqlalchemy.orm import Session, joinedload

from app.models.category import Category
from app.models.question import Question
from app.service.question import question_service


class QuestionBankService:
    def generate(
        self,
        db: Session,
        question_count: int,
        difficulty: Optional[int] = None,
        question_types: Optional[List[str]] = None,
        category_id: Optional[str] = None,
        category_ids: Optional[List[str]] = None,
        is_superuser: bool = False,
        fixed_question_ids: Optional[List[str]] = None,
    ) -> List[Question]:
        if fixed_question_ids:
            return self._load_fixed_questions(
                db=db,
                question_ids=fixed_question_ids,
                is_superuser=is_superuser,
            )

        query = db.query(Question).options(
            joinedload(Question.options),
            joinedload(Question.knowledge_tags),
        )

        if not is_superuser:
            query = query.filter(Question.review_status == question_service.REVIEW_APPROVED)

        if difficulty:
            query = query.filter(Question.difficulty == difficulty)

        if question_types:
            query = query.filter(Question.type.in_(question_types))

        if category_ids:
            query = query.join(Question.knowledge_tags).filter(Category.id.in_(category_ids))
        elif category_id:
            query = query.join(Question.knowledge_tags).filter(Category.id == category_id)

        pool = query.distinct().all()
        random.shuffle(pool)
        return pool[:question_count]

    def generate_by_type_counts(
        self,
        db: Session,
        type_counts: Dict[str, int],
        knowledge_tag_ids: Optional[List[str]] = None,
        difficulty: Optional[int] = None,
        is_superuser: bool = False,
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
                .filter(Question.type == question_type)
            )

            if not is_superuser:
                query = query.filter(Question.review_status == question_service.REVIEW_APPROVED)

            if difficulty:
                query = query.filter(Question.difficulty == difficulty)

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

    def _load_fixed_questions(
        self,
        db: Session,
        question_ids: List[str],
        is_superuser: bool = False,
    ) -> List[Question]:
        if not question_ids:
            return []

        query = (
            db.query(Question)
            .options(joinedload(Question.options), joinedload(Question.knowledge_tags))
            .filter(Question.id.in_(question_ids))
        )
        if not is_superuser:
            query = query.filter(Question.review_status == question_service.REVIEW_APPROVED)

        questions = query.all()
        question_map = {item.id: item for item in questions}
        ordered = [question_map[qid] for qid in question_ids if qid in question_map]
        return ordered


question_bank_service = QuestionBankService()
