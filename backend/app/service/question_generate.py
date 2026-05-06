import json
from typing import List, Optional, AsyncGenerator

import httpx
from pydantic import BaseModel

from app.config import get_settings


class GeneratedOption(BaseModel):
    option_key: str
    content: str
    is_correct: bool


class GeneratedQuestion(BaseModel):
    title: str
    type: str
    difficulty: int
    answer: str
    explanation: str
    options: Optional[List[GeneratedOption]] = None


class QuestionGenerateService:
    def __init__(self):
        self.settings = get_settings()

    async def _call_llm_stream(
        self, prompt: str
    ) -> AsyncGenerator[str, None]:
        llm_url = self.settings.LLM_API_URL
        llm_key = self.settings.LLM_API_KEY
        llm_model = self.settings.LLM_MODEL

        headers = {
            "Authorization": f"Bearer {llm_key}",
            "Content-Type": "application/json",
        }

        payload = {
            "model": llm_model,
            "messages": [{"role": "user", "content": prompt}],
            "temperature": 0.5,
            "stream": True,
            "response_format": {"type": "json_object"},
        }

        async with httpx.AsyncClient(timeout=300.0) as client:
            async with client.stream(
                "POST",
                f"{llm_url}/chat/completions",
                headers=headers,
                json=payload,
            ) as response:
                response.raise_for_status()
                async for line in response.aiter_lines():
                    if line.startswith("data: "):
                        data_str = line[6:]
                        if data_str == "[DONE]":
                            break
                        try:
                            data = json.loads(data_str)
                            content = data.get("choices", [{}])[0].get("delta", {}).get("content", "")
                            if content:
                                yield content
                        except json.JSONDecodeError:
                            continue

    def _build_prompt(
        self,
        knowledge_items: List[dict],
        question_type: str,
        difficulty: int,
    ) -> str:
        type_desc = {
            "single_choice": "单选题",
            "multi_choice": "多选题",
            "true_false": "判断题",
            "fill_blank": "填空题",
            "essay": "论述题",
            "short_answer": "简答题",
        }
        type_name = type_desc.get(question_type, question_type)

        knowledge_text = "\n".join([
            f"- {item['name']}" + (f": {item['description']}" if item.get('description') else "")
            for item in knowledge_items
        ])

        return f"""你是专业教师，请严格根据下方的知识点内容，生成一道{type_name}。

【绝对要求】：
1. 题目100%围绕以下知识点，不得出与这些知识点无关的题目
2. 题目应直接考察学生对这些知识点的理解和应用能力
3. 难度：{difficulty}（1简单-5困难）
4. 只输出JSON，不要其他文字

【知识点详细内容】：
{knowledge_text}

如果知识点描述为空，请基于知识点名称，结合该领域的常识进行出题。

【输出JSON格式】：
{{
    "title": "题目内容/题干",
    "type": "{question_type}",
    "difficulty": {difficulty},
    "answer": "正确答案",
    "explanation": "题目解析",
    "options": [
        {{"option_key": "A", "content": "选项A内容", "is_correct": false}},
        {{"option_key": "B", "content": "选项B内容", "is_correct": false}},
        {{"option_key": "C", "content": "选项C内容", "is_correct": false}},
        {{"option_key": "D", "content": "选项D内容", "is_correct": false}}
    ]
}}

注意：
- 如果是选择题，需要提供4个选项，并标记正确答案
- 如果是判断题，is_correct为true表示正确，false表示错误
- 如果是填空题或简答题，不需要options字段
- 确保题目内容准确、清晰、符合难度要求
- 题目必须围绕上述知识点的具体内容"""

    def _parse_response(self, content: str) -> GeneratedQuestion:
        data = json.loads(content)
        options = None
        if data.get("options"):
            options = [
                GeneratedOption(
                    option_key=opt["option_key"],
                    content=opt["content"],
                    is_correct=opt.get("is_correct", False),
                )
                for opt in data["options"]
            ]
        return GeneratedQuestion(
            title=data.get("title", ""),
            type=data.get("type", "single_choice"),
            difficulty=data.get("difficulty", 3),
            answer=data.get("answer", ""),
            explanation=data.get("explanation", ""),
            options=options,
        )

    async def generate_question_stream(
        self,
        knowledge_items: List[dict],
        question_type: str,
        difficulty: int,
    ):
        prompt = self._build_prompt(knowledge_items, question_type, difficulty)
        full_content = ""

        async for chunk in self._call_llm_stream(prompt):
            full_content += chunk
            yield f"data: {json.dumps({'type': 'chunk', 'content': chunk})}\n\n"

        try:
            question = self._parse_response(full_content)
            yield f"data: {json.dumps({'type': 'done', 'data': question.model_dump()})}\n\n"
        except Exception as e:
            yield f"data: {json.dumps({'type': 'error', 'message': str(e)})}\n\n"


question_generate_service = QuestionGenerateService()