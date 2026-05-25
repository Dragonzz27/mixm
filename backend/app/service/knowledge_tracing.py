from dataclasses import dataclass
from typing import Dict, Iterable, List

from sqlalchemy.orm import Session

from app.models.knowledge_tracing import KnowledgeMastery, SimulationAnswerTrace


@dataclass
class TraceItem:
    # 单次作答的最小追踪单元：题目、知识点、选择、对错。
    question_id: str
    knowledge_tag: str
    selected_index: int | None
    correct_index: int
    is_correct: bool


class KnowledgeTracingService:
    def __init__(self, alpha: float = 0.18, beta: float = 0.22):
        self.alpha = alpha
        self.beta = beta

    def record_and_update(
        self,
        db: Session,
        user_id: str,
        trace_items: Iterable[TraceItem],
    ) -> Dict[str, float]:
        # 逐条写入作答轨迹，并同步更新每个知识点的掌握度。
        latest_mastery: Dict[str, float] = {}

        for item in trace_items:
            db.add(
                SimulationAnswerTrace(
                    user_id=user_id,
                    question_id=item.question_id,
                    knowledge_tag=item.knowledge_tag,
                    selected_index=item.selected_index,
                    correct_index=item.correct_index,
                    is_correct=item.is_correct,
                )
            )

            mastery_row = (
                db.query(KnowledgeMastery)
                .filter(
                    KnowledgeMastery.user_id == user_id,
                    KnowledgeMastery.knowledge_tag == item.knowledge_tag,
                )
                .first()
            )
            if not mastery_row:
                # 首次出现的知识点默认从 0.5 开始，避免冷启动全 0。
                mastery_row = KnowledgeMastery(
                    user_id=user_id,
                    knowledge_tag=item.knowledge_tag,
                    mastery=0.5,
                    attempts=0,
                    correct_attempts=0,
                )
                db.add(mastery_row)

            p = mastery_row.mastery
            if item.is_correct:
                # 答对时缓慢上升，越接近 1 提升越小。
                p = p + self.alpha * (1.0 - p)
                mastery_row.correct_attempts += 1
            else:
                # 答错时按比例回落，保持平滑而不是直接归零。
                p = p - self.beta * p
            mastery_row.mastery = max(0.0, min(1.0, p))
            mastery_row.attempts += 1
            latest_mastery[item.knowledge_tag] = round(mastery_row.mastery, 4)

        db.commit()
        return latest_mastery

    def get_user_mastery(self, db: Session, user_id: str) -> List[KnowledgeMastery]:
        # 返回某个学生当前的全部知识点掌握度，供前端展示。
        return (
            db.query(KnowledgeMastery)
            .filter(KnowledgeMastery.user_id == user_id)
            .order_by(KnowledgeMastery.mastery.asc())
            .all()
        )


knowledge_tracing_service = KnowledgeTracingService()
