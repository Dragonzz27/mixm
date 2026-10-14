import uuid

from sqlalchemy import Boolean, Column, DateTime, Float, Integer, String, Text, func

from app.database import Base


class KnowledgeMastery(Base):
    __tablename__ = "knowledge_mastery"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), index=True, nullable=False)
    knowledge_tag = Column(String(200), index=True, nullable=False)
    mastery = Column(Float, nullable=False, default=0.5)
    attempts = Column(Integer, nullable=False, default=0)
    correct_attempts = Column(Integer, nullable=False, default=0)
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())


class SimulationAnswerTrace(Base):
    __tablename__ = "simulation_answer_trace"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), index=True, nullable=False)
    question_id = Column(String(36), index=True, nullable=False)
    knowledge_tag = Column(String(200), index=True, nullable=False)
    selected_index = Column(Integer, nullable=True)
    correct_index = Column(Integer, nullable=False)
    is_correct = Column(Boolean, nullable=False, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())


class SimulationAttempt(Base):
    __tablename__ = "simulation_attempt"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), index=True, nullable=False)
    status = Column(String(20), index=True, nullable=False, default="in_progress")
    paper_source = Column(String(20), nullable=False, default="random")
    practice_id = Column(String(36), nullable=True)
    questions_payload = Column(Text, nullable=False, default="[]")
    answers_payload = Column(Text, nullable=False, default="[]")
    result_payload = Column(Text, nullable=True)
    started_at = Column(DateTime(timezone=True), server_default=func.now())
    submitted_at = Column(DateTime(timezone=True), nullable=True)
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())
