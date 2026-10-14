from typing import Dict, List, Optional

from pydantic import BaseModel, Field


class SimulationGenerateRequest(BaseModel):
    course_id: Optional[str] = Field(default=None, description="课程/分类ID")
    question_count: Optional[int] = Field(default=None, ge=1, le=50)
    difficulty: Optional[int] = Field(default=None, ge=1, le=5, description="难度等级1-5")
    question_types: Optional[List[str]] = Field(default=None, description="题型")
    type_counts: Optional[Dict[str, int]] = Field(default=None, description="按题型配置数量")
    time_limit_minutes: Optional[int] = Field(default=None, ge=1, le=180)
    generation_mode: Optional[str] = Field(
        default="random",
        description="出题模式：random(题库随机) / ai(智能体生成) / practice(练习卷固定题单)",
    )
    practice_id: Optional[str] = Field(default=None, description="练习卷ID，仅 practice 模式使用")


class SimulationQuestion(BaseModel):
    id: str
    stem: str
    options: List[str]
    correct_index: int
    difficulty: int = 3
    score: float = 0.0
    correct_indices: Optional[List[int]] = None
    correct_answer: Optional[str] = None
    knowledge_tag: Optional[str] = None
    knowledge_tags: List[str] = Field(..., min_items=1)
    question_type: str = "single_choice"


class SimulationGenerateResponse(BaseModel):
    attempt_id: Optional[str] = None
    attempt_status: str = "in_progress"
    settings: SimulationGenerateRequest
    questions: List[SimulationQuestion]


class SimulationAnswerItem(BaseModel):
    question_id: str
    selected_index: Optional[int] = None
    answer_text: Optional[str] = None


class SimulationSubmitRequest(BaseModel):
    attempt_id: Optional[str] = None
    questions: List[SimulationQuestion]
    answers: List[SimulationAnswerItem]


class SimulationSubmitResponse(BaseModel):
    attempt_id: Optional[str] = None
    attempt_status: str = "submitted"
    total_score: float
    score: float
    total_questions: int
    answered_questions: int
    correct_count: int
    accuracy: float
    weak_knowledge_points: List[str]
    knowledge_accuracy: Dict[str, float]
    knowledge_mastery: Dict[str, float] = Field(default_factory=dict)
    q_matrix: List[Dict[str, object]] = Field(default_factory=list)
    q_matrix_by_question: Dict[str, List[str]] = Field(default_factory=dict)
    q_matrix_by_tag: Dict[str, List[str]] = Field(default_factory=dict)


class SimulationAnswerReviewResponse(BaseModel):
    attempt_id: str
    attempt_status: str
    questions: List[SimulationQuestion] = Field(default_factory=list)
    answers: List[SimulationAnswerItem] = Field(default_factory=list)
    result: Optional[SimulationSubmitResponse] = None


class SimulationTraceItem(BaseModel):
    id: str
    user_id: str
    question_id: str
    knowledge_tag: str
    selected_index: Optional[int] = None
    correct_index: int
    is_correct: bool
    created_at: Optional[str] = None


class SimulationTraceListResponse(BaseModel):
    total: int
    items: List[SimulationTraceItem] = Field(default_factory=list)
