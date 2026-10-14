import logging
from dataclasses import dataclass
from typing import Dict, Iterable, List, Optional

import httpx
from sqlalchemy.orm import Session

from app.config import get_settings
from app.models.knowledge_tracing import KnowledgeMastery, SimulationAnswerTrace

logger = logging.getLogger(__name__)


@dataclass
class TraceItem:
    question_id: str
    knowledge_tag: str
    selected_index: Optional[int]
    correct_index: int
    is_correct: bool


class KnowledgeTracingService:
    def __init__(self, alpha: float = 0.18, beta: float = 0.22):
        self.alpha = alpha
        self.beta = beta
        settings = get_settings()
        self.backend = (settings.KT_MODEL_BACKEND or "heuristic").lower()
        self.kt_service_url = settings.KT_SERVICE_URL
        self.kt_timeout = settings.KT_SERVICE_TIMEOUT_SECONDS

    def _record_trace_rows(self, db: Session, user_id: str, trace_items: List[TraceItem]) -> None:
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
        db.flush()

    def _batch_stats(self, trace_items: List[TraceItem]) -> Dict[str, Dict[str, int]]:
        stats: Dict[str, Dict[str, int]] = {}
        for item in trace_items:
            tag = item.knowledge_tag
            stats.setdefault(tag, {"attempts": 0, "correct_attempts": 0})
            stats[tag]["attempts"] += 1
            if item.is_correct:
                stats[tag]["correct_attempts"] += 1
        return stats

    def _upsert_mastery_rows(
        self,
        db: Session,
        user_id: str,
        mastery_by_tag: Dict[str, float],
        stats_by_tag: Dict[str, Dict[str, int]],
    ) -> Dict[str, float]:
        latest: Dict[str, float] = {}
        for tag, mastery in mastery_by_tag.items():
            row = (
                db.query(KnowledgeMastery)
                .filter(
                    KnowledgeMastery.user_id == user_id,
                    KnowledgeMastery.knowledge_tag == tag,
                )
                .first()
            )
            if not row:
                row = KnowledgeMastery(
                    user_id=user_id,
                    knowledge_tag=tag,
                    mastery=0.5,
                    attempts=0,
                    correct_attempts=0,
                )
                db.add(row)

            increments = stats_by_tag.get(tag, {"attempts": 0, "correct_attempts": 0})
            row.attempts += int(increments.get("attempts", 0))
            row.correct_attempts += int(increments.get("correct_attempts", 0))
            row.mastery = max(0.0, min(1.0, float(mastery)))
            latest[tag] = round(row.mastery, 4)
        return latest

    def _heuristic_mastery(self, db: Session, user_id: str, trace_items: List[TraceItem]) -> Dict[str, float]:
        latest_mastery: Dict[str, float] = {}

        for item in trace_items:
            row = (
                db.query(KnowledgeMastery)
                .filter(
                    KnowledgeMastery.user_id == user_id,
                    KnowledgeMastery.knowledge_tag == item.knowledge_tag,
                )
                .first()
            )
            if not row:
                row = KnowledgeMastery(
                    user_id=user_id,
                    knowledge_tag=item.knowledge_tag,
                    mastery=0.5,
                    attempts=0,
                    correct_attempts=0,
                )
                db.add(row)

            p = row.mastery
            if item.is_correct:
                p = p + self.alpha * (1.0 - p)
                row.correct_attempts += 1
            else:
                p = p - self.beta * p
            row.mastery = max(0.0, min(1.0, p))
            row.attempts += 1
            latest_mastery[item.knowledge_tag] = round(row.mastery, 4)

        return latest_mastery

    def _call_kt_service(
        self,
        user_id: str,
        trace_items: List[TraceItem],
        prev_mastery: Dict[str, float],
    ) -> Dict[str, float]:
        payload = {
            "user_id": user_id,
            "traces": [
                {
                    "question_id": t.question_id,
                    "knowledge_tag": t.knowledge_tag,
                    "is_correct": bool(t.is_correct),
                }
                for t in trace_items
            ],
            "prev_mastery": prev_mastery,
        }
        with httpx.Client(timeout=self.kt_timeout) as client:
            resp = client.post(f"{self.kt_service_url.rstrip('/')}/predict", json=payload)
            resp.raise_for_status()
            data = resp.json()
            mastery = data.get("mastery_by_tag") or {}
            return {k: float(v) for k, v in mastery.items()}

    def record_and_update(
        self,
        db: Session,
        user_id: str,
        trace_items: Iterable[TraceItem],
    ) -> Dict[str, float]:
        items = list(trace_items)
        if not items:
            return {}

        # 1) Always persist answer trace first.
        self._record_trace_rows(db, user_id, items)

        # 2) Then compute/update mastery.
        if self.backend == "kt_service":
            prev_mastery: Dict[str, float] = {}
            for tag in {i.knowledge_tag for i in items}:
                row = (
                    db.query(KnowledgeMastery)
                    .filter(
                        KnowledgeMastery.user_id == user_id,
                        KnowledgeMastery.knowledge_tag == tag,
                    )
                    .first()
                )
                prev_mastery[tag] = float(row.mastery) if row else 0.5

            try:
                predicted = self._call_kt_service(user_id, items, prev_mastery)
                stats = self._batch_stats(items)
                latest = self._upsert_mastery_rows(db, user_id, predicted, stats)
                db.commit()
                return latest
            except Exception as exc:
                logger.warning("KT service failed, fallback to heuristic updater: %s", exc)
                latest = self._heuristic_mastery(db, user_id, items)
                db.commit()
                return latest

        latest = self._heuristic_mastery(db, user_id, items)
        db.commit()
        return latest

    def get_user_mastery(self, db: Session, user_id: str) -> List[KnowledgeMastery]:
        return (
            db.query(KnowledgeMastery)
            .filter(KnowledgeMastery.user_id == user_id)
            .order_by(KnowledgeMastery.mastery.asc())
            .all()
        )

    def get_backend_status(self) -> Dict[str, object]:
        return {
            "backend": self.backend,
            "kt_service_url": self.kt_service_url if self.backend == "kt_service" else None,
        }


knowledge_tracing_service = KnowledgeTracingService()
