import json
import random
import uuid
from typing import Dict, List, Optional

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session, joinedload

from app.database import get_db
from app.models.category import Category
from app.models.content import Content
from app.models.content_category import ContentCategory
from app.models.practice import Practice, PracticeQuestion
from app.models.question import Question
from app.schema.simulation import (
    SimulationGenerateRequest,
    SimulationGenerateResponse,
    SimulationQuestion,
    SimulationSubmitRequest,
    SimulationSubmitResponse,
)
from app.service.auth import AuthService
from app.service.knowledge_tracing import TraceItem, knowledge_tracing_service
from app.service.question_bank import question_bank_service
from app.service.question_generate import question_generate_service

router = APIRouter()
get_current_user = AuthService().get_current_user


def _collect_descendant_ids(db: Session, root_id: str) -> List[str]:
    ids = [root_id]
    queue = [root_id]
    while queue:
        current = queue.pop(0)
        children = (
            db.query(Category.id)
            .filter(Category.parent_id == current, Category.is_active == True)
            .all()
        )
        child_ids = [row[0] for row in children]
        ids.extend(child_ids)
        queue.extend(child_ids)
    return ids


def _normalize_question_type(question_type: str) -> str:
    allowed = {
        "single_choice",
        "multi_choice",
        "true_false",
        "fill_blank",
        "short_answer",
    }
    return question_type if question_type in allowed else "single_choice"


def _build_options(correct_text: str, distractors: List[str]) -> List[str]:
    options = [correct_text] + distractors
    random.shuffle(options)
    return options


def _pick_question_type(question_types: Optional[List[str]], index: int) -> str:
    if not question_types:
        return "single_choice"
    normalized = [_normalize_question_type(qt) for qt in question_types]
    if not normalized:
        return "single_choice"
    return normalized[index % len(normalized)]


def _map_question_type_from_bank(question_type: str) -> str:
    mapping = {
        "blank": "fill_blank",
        "essay": "short_answer",
    }
    return mapping.get(question_type, question_type)


def _map_question_type_to_bank(question_type: str) -> str:
    mapping = {
        "fill_blank": "blank",
        "short_answer": "essay",
    }
    return mapping.get(question_type, question_type)


def _map_question_types_to_bank(question_types: Optional[List[str]]) -> Optional[List[str]]:
    if not question_types:
        return None
    mapped = [_map_question_type_to_bank(item) for item in question_types if item]
    return list(dict.fromkeys(mapped)) if mapped else None


def _fallback_knowledge_tags(content: Content, category: Optional[Category]) -> List[str]:
    if content.title:
        return [content.title]
    if category and category.name:
        return [category.name]
    return ["未分类"]


def _build_content_knowledge_map(
    db: Session,
    rows: List[tuple[Content, str]],
) -> Dict[str, List[str]]:
    knowledge_map: Dict[str, List[str]] = {}
    category_ids = list({cat_id for _, cat_id in rows if cat_id})
    if not category_ids:
        return knowledge_map

    cats = db.query(Category.id, Category.name, Category.parent_id).filter(Category.id.in_(category_ids)).all()
    category_name_map = {cid: name for cid, name, _ in cats}
    leaf_name_cache: Dict[str, List[str]] = {}

    def get_leaf_names(cat_id: str) -> List[str]:
        if cat_id in leaf_name_cache:
            return leaf_name_cache[cat_id]
        descendants = _collect_descendant_ids(db, cat_id)
        if not descendants:
            leaf_name_cache[cat_id] = []
            return []

        descendant_rows = (
            db.query(Category.id, Category.name)
            .filter(Category.id.in_(descendants), Category.is_active == True)
            .all()
        )
        descendant_ids = [cid for cid, _ in descendant_rows]
        parent_rows = (
            db.query(Category.parent_id)
            .filter(Category.parent_id.in_(descendant_ids), Category.is_active == True)
            .all()
        )
        parent_set = {pid for (pid,) in parent_rows if pid}
        leaf_names = [name for cid, name in descendant_rows if cid not in parent_set and name]
        leaf_name_cache[cat_id] = leaf_names
        return leaf_names

    for content, cat_id in rows:
        if not cat_id:
            continue
        knowledge_map.setdefault(content.id, [])
        leaf_names = get_leaf_names(cat_id)
        if leaf_names:
            for tag in leaf_names:
                if tag not in knowledge_map[content.id]:
                    knowledge_map[content.id].append(tag)
            continue
        # fallback to current category name when no leaf is available
        tag = category_name_map.get(cat_id)
        if tag and tag not in knowledge_map[content.id]:
            knowledge_map[content.id].append(tag)
    return knowledge_map


def _normalize_text_answer(value: Optional[str]) -> str:
    if not value:
        return ""
    return "".join(value.lower().split())


def _build_random_questions(
    payload: SimulationGenerateRequest,
    contents: List[Content],
    category: Optional[Category],
    knowledge_map: Dict[str, List[str]],
) -> List[SimulationQuestion]:
    titles = [content.title for content in contents if content.title]
    random.shuffle(titles)

    question_count = payload.question_count or 10
    questions: List[SimulationQuestion] = []

    for idx in range(question_count):
        content = contents[idx % len(contents)]
        correct_text = content.title
        distractors = [title for title in titles if title != correct_text]
        if len(distractors) < 3:
            distractors.extend(["骞叉壈椤笰", "骞叉壈椤笲", "骞叉壈椤笴"])
        distractors = distractors[:3]
        question_type = _pick_question_type(payload.question_types, idx)
        if question_type == "true_false":
            options = ["姝ｇ‘", "閿欒"]
            correct_index = 0
            correct_answer = "姝ｇ‘"
        elif question_type in {"fill_blank", "short_answer"}:
            options = []
            correct_index = 0
            correct_answer = correct_text
        else:
            options = _build_options(correct_text, distractors)
            correct_index = options.index(correct_text)
            correct_answer = None
        knowledge_tags = knowledge_map.get(content.id) or _fallback_knowledge_tags(content, category)
        stem = f"以下哪一项最符合“{correct_text}”？"
        if payload.difficulty:
            stem = f"(闅惧害{payload.difficulty}) {stem}"
        if question_type in {"fill_blank", "short_answer"}:
            stem = f"请填写与“{correct_text}”相关的关键内容。"
        questions.append(
            SimulationQuestion(
                id=str(uuid.uuid4()),
                stem=stem,
                options=options,
                correct_index=correct_index,
                correct_answer=correct_answer,
                knowledge_tag=" / ".join(knowledge_tags[:3]),
                knowledge_tags=knowledge_tags,
                question_type=question_type,
            )
        )
    return questions


async def _build_ai_questions(
    payload: SimulationGenerateRequest,
    contents: List[Content],
    category: Optional[Category],
    knowledge_map: Dict[str, List[str]],
) -> List[SimulationQuestion]:
    question_count = payload.question_count or 10
    difficulty_level = payload.difficulty or 3
    questions: List[SimulationQuestion] = []
    for idx in range(question_count):
        content = contents[idx % len(contents)]
        knowledge_tags = knowledge_map.get(content.id) or _fallback_knowledge_tags(content, category)
        knowledge_items = [
            {"name": kp, "description": content.summary or content.title or ""}
            for kp in knowledge_tags
        ]

        question_type = _pick_question_type(payload.question_types, idx)
        prompt = question_generate_service._build_prompt(
            knowledge_items=knowledge_items,
            question_type=question_type,
            difficulty=difficulty_level,
        )
        full_content = ""
        async for chunk in question_generate_service._call_llm_stream(prompt):
            full_content += chunk
        generated = question_generate_service._parse_response(full_content)

        options = [opt.content for opt in (generated.options or []) if opt.content]
        correct_index = 0
        if generated.options:
            for i, opt in enumerate(generated.options):
                if opt.is_correct:
                    correct_index = i
                    break
        if not options:
            if question_type == "true_false":
                options = ["姝ｇ‘", "閿欒"]
                correct_index = 0
            elif question_type in {"fill_blank", "short_answer"}:
                options = []
                correct_index = 0
            else:
                answer_text = generated.answer or "正确答案"
                options = _build_options(answer_text, ["干扰项A", "干扰项B", "干扰项C"])
                correct_index = options.index(answer_text)

        correct_answer = None
        if question_type in {"fill_blank", "short_answer"}:
            correct_answer = generated.answer or content.title or "参考答案"

        stem = generated.title or f"请回答与“{content.title}”相关的问题"
        questions.append(
            SimulationQuestion(
                id=str(uuid.uuid4()),
                stem=stem,
                options=options,
                correct_index=correct_index,
                correct_answer=correct_answer,
                knowledge_tag=" / ".join(knowledge_tags[:3]),
                knowledge_tags=knowledge_tags,
                question_type=question_type,
            )
        )
    return questions


def _build_bank_questions(
    payload: SimulationGenerateRequest,
    db: Session,
    category: Optional[Category],
    current_user,
) -> List[SimulationQuestion]:
    if payload.type_counts:
        type_counts_bank: Dict[str, int] = {}
        for q_type, count in payload.type_counts.items():
            if not q_type:
                continue
            mapped = _map_question_type_to_bank(q_type)
            type_counts_bank[mapped] = type_counts_bank.get(mapped, 0) + max(int(count or 0), 0)
        category_ids = _collect_descendant_ids(db, category.id) if category else None
        selected, shortages = question_bank_service.generate_by_type_counts(
            db=db,
            type_counts=type_counts_bank,
            knowledge_tag_ids=category_ids,
            difficulty=payload.difficulty,
            is_superuser=bool(getattr(current_user, "is_superuser", False)),
        )
        if not selected:
            raise HTTPException(status_code=400, detail="题库中暂无符合条件的题目")
        if shortages:
            shortage_text = "，".join([f"{k}:缺{v}" for k, v in shortages.items()])
            raise HTTPException(status_code=400, detail=f"题目数量不足：{shortage_text}")
        return [_to_simulation_question(item) for item in selected]

    question_count = payload.question_count or 10
    bank_types = _map_question_types_to_bank(payload.question_types)
    category_ids = _collect_descendant_ids(db, category.id) if category else None
    selected = question_bank_service.generate(
        db=db,
        question_count=question_count,
        difficulty=payload.difficulty,
        question_types=bank_types,
        category_id=category.id if category else None,
        category_ids=category_ids,
        is_superuser=bool(getattr(current_user, "is_superuser", False)),
    )
    if not selected:
        raise HTTPException(status_code=400, detail="题库中暂无符合条件的题目")
    return [_to_simulation_question(item) for item in selected]


def _to_simulation_question(question: Question) -> SimulationQuestion:
    question_type = _map_question_type_from_bank(question.type)
    options = []
    correct_indices: List[int] = []
    correct_index = 0

    if question.options:
        sorted_options = sorted(question.options, key=lambda opt: opt.option_key)
        options = [opt.content for opt in sorted_options]
        correct_indices = [idx for idx, opt in enumerate(sorted_options) if opt.is_correct]
        if correct_indices:
            correct_index = correct_indices[0]

    knowledge_tags = [tag.name for tag in (question.knowledge_tags or []) if tag.name]
    if not knowledge_tags and question.tags:
        knowledge_tags = list(question.tags)
    if not knowledge_tags:
        knowledge_tags = ["未分类"]

    return SimulationQuestion(
        id=question.id,
        stem=question.title,
        options=options,
        correct_index=correct_index,
        correct_indices=correct_indices or None,
        correct_answer=question.answer if question_type in {"fill_blank", "short_answer"} else None,
        knowledge_tag=" / ".join(knowledge_tags[:3]),
        knowledge_tags=knowledge_tags,
        question_type=question_type,
    )


def _build_practice_questions(
    payload: SimulationGenerateRequest,
    db: Session,
    current_user,
) -> List[SimulationQuestion]:
    if not payload.practice_id:
        raise HTTPException(status_code=400, detail="practice 模式需要 practice_id")

    practice = (
        db.query(Practice)
        .options(joinedload(Practice.practice_questions).joinedload(PracticeQuestion.question))
        .filter(Practice.id == payload.practice_id)
        .first()
    )
    if not practice:
        raise HTTPException(status_code=404, detail="练习不存在")

    question_ids = [item.question_id for item in (practice.practice_questions or []) if item.question_id]
    if not question_ids:
        raise HTTPException(status_code=400, detail="该练习还没有题目")

    selected = question_bank_service.generate(
        db=db,
        question_count=payload.question_count or len(question_ids),
        is_superuser=bool(getattr(current_user, "is_superuser", False)),
        fixed_question_ids=question_ids,
    )
    if payload.question_count:
        selected = selected[: payload.question_count]
    if not selected:
        raise HTTPException(status_code=400, detail="该练习下暂无可作答题目")
    return [_to_simulation_question(item) for item in selected]
@router.post("/generate", response_model=SimulationGenerateResponse)
async def generate_simulation(
    payload: SimulationGenerateRequest,
    db: Session = Depends(get_db),
    current_user=Depends(get_current_user),
):
    category: Optional[Category] = None
    if payload.course_id:
        category = (
            db.query(Category)
            .filter(Category.id == payload.course_id, Category.is_active == True)
            .first()
        )
        if not category:
            raise HTTPException(status_code=404, detail="Course category not found")

    mode = (payload.generation_mode or "random").lower()
    if mode == "ai":
        category_ids = _collect_descendant_ids(db, category.id) if category else []
        contents: List[Content] = []
        knowledge_map: Dict[str, List[str]] = {}

        if category_ids:
            rows = (
                db.query(Content, ContentCategory.category_id)
                .join(ContentCategory, Content.id == ContentCategory.content_id)
                .filter(
                    ContentCategory.category_id.in_(category_ids),
                    Content.is_active == True,
                    Content.status == "published",
                )
                .all()
            )
            contents = [content for content, _ in rows]
            knowledge_map = _build_content_knowledge_map(db, rows)

        if not contents:
            contents = (
                db.query(Content)
                .filter(Content.is_active == True, Content.status == "published")
                .limit(100)
                .all()
            )

        if not contents:
            raise HTTPException(status_code=400, detail="No published content available")
        try:
            questions = await _build_ai_questions(payload, contents, category, knowledge_map)
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"鏅鸿兘浣撶敓鎴愬け璐? {str(e)}")
    elif mode == "practice":
        questions = _build_practice_questions(payload, db, current_user)
    else:
        questions = _build_bank_questions(payload, db, category, current_user)

    return SimulationGenerateResponse(settings=payload, questions=questions)


@router.post("/submit", response_model=SimulationSubmitResponse)
def submit_simulation(
    payload: SimulationSubmitRequest,
    current_user=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    answer_map = {item.question_id: item.selected_index for item in payload.answers}
    answer_text_map = {item.question_id: item.answer_text for item in payload.answers}
    total = len(payload.questions)
    if total == 0:
        raise HTTPException(status_code=400, detail="Questions are required")

    answered = 0
    correct = 0
    knowledge_stats: Dict[str, Dict[str, int]] = {}
    trace_items: List[TraceItem] = []
    q_matrix: List[Dict[str, object]] = []
    q_matrix_by_question: Dict[str, List[str]] = {}
    q_matrix_by_tag: Dict[str, List[str]] = {}

    for question in payload.questions:
        selected = answer_map.get(question.id)
        answer_text = answer_text_map.get(question.id)
        is_text_type = question.question_type in {"fill_blank", "short_answer"}

        if is_text_type:
            normalized = _normalize_text_answer(answer_text)
            if normalized:
                answered += 1
            expected = _normalize_text_answer(question.correct_answer)
            is_correct = bool(normalized) and normalized == expected
        elif question.question_type == "multi_choice":
            picks: List[int] = []
            if answer_text:
                try:
                    raw = json.loads(answer_text)
                    if isinstance(raw, list):
                        picks = [int(item) for item in raw]
                except Exception:
                    picks = []
            if not picks and selected is not None:
                picks = [selected]
            if picks:
                answered += 1
            expected_indices = set(question.correct_indices or [question.correct_index])
            is_correct = set(picks) == expected_indices
        else:
            if selected is not None:
                answered += 1
            is_correct = selected == question.correct_index if selected is not None else False
        if is_correct:
            correct += 1

        q_tags = question.knowledge_tags or ([question.knowledge_tag] if question.knowledge_tag else [])
        if not q_tags:
            raise HTTPException(
                status_code=400,
                detail=f"Question {question.id} missing knowledge_tags",
            )
        q_matrix_by_question[question.id] = list(dict.fromkeys(q_tags))
        for kp in q_tags:
            knowledge_stats.setdefault(kp, {"total": 0, "correct": 0})
            knowledge_stats[kp]["total"] += 1
            if is_correct:
                knowledge_stats[kp]["correct"] += 1
            q_matrix_by_tag.setdefault(kp, [])
            if question.id not in q_matrix_by_tag[kp]:
                q_matrix_by_tag[kp].append(question.id)

            trace_items.append(
                TraceItem(
                    question_id=question.id,
                    knowledge_tag=kp,
                    selected_index=selected,
                    correct_index=question.correct_index,
                    is_correct=is_correct,
                )
            )
            q_matrix.append(
                {
                    "question_id": question.id,
                    "knowledge_tag": kp,
                    "selected_index": selected,
                    "correct_index": question.correct_index,
                    "is_correct": is_correct,
                }
            )

    accuracy = round((correct / total) * 100, 2) if total > 0 else 0.0
    knowledge_accuracy: Dict[str, float] = {}
    weak_points: List[str] = []
    for kp, stat in knowledge_stats.items():
        kp_accuracy = round((stat["correct"] / stat["total"]) * 100, 2) if stat["total"] > 0 else 0.0
        knowledge_accuracy[kp] = kp_accuracy
        if kp_accuracy < 60:
            weak_points.append(kp)
    weak_points.sort(key=lambda name: knowledge_accuracy.get(name, 0.0))

    knowledge_mastery = knowledge_tracing_service.record_and_update(
        db=db,
        user_id=current_user.id,
        trace_items=trace_items,
    )

    return SimulationSubmitResponse(
        total_questions=total,
        answered_questions=answered,
        correct_count=correct,
        accuracy=accuracy,
        weak_knowledge_points=weak_points,
        knowledge_accuracy=knowledge_accuracy,
        knowledge_mastery=knowledge_mastery,
        q_matrix=q_matrix,
        q_matrix_by_question=q_matrix_by_question,
        q_matrix_by_tag=q_matrix_by_tag,
    )
