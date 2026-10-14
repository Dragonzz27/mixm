import json
from typing import Optional, AsyncGenerator

import httpx
from sqlalchemy.orm import Session

from app.config import get_settings
from app.models.category import Category
from app.schema.knowledge_tree import (
    KnowledgeTree,
    Chapter,
    Section,
    KnowledgePoint,
    KnowledgeElement,
    OutlineAnalysisRequest,
    OutlineAnalysisResponse,
)


class KnowledgeTreeService:
    def __init__(self):
        self.settings = get_settings()

    async def _call_llm_stream(self, prompt: str) -> AsyncGenerator[str, None]:
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
            "temperature": 0.3,
            "stream": True,
            "response_format": {"type": "json_object"},
        }

        async with httpx.AsyncClient(timeout=360.0) as client:
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

    async def _call_llm(self, prompt: str) -> str:
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
            "temperature": 0.3,
            "response_format": {"type": "json_object"},
        }

        async with httpx.AsyncClient(timeout=120.0) as client:
            response = await client.post(
                f"{llm_url}/chat/completions",
                headers=headers,
                json=payload,
            )
            response.raise_for_status()
            result = response.json()
            return result["choices"][0]["message"]["content"]

    def _build_prompt(self, request: OutlineAnalysisRequest) -> str:
        course_name = request.course_name or "课程"
        return f"""请分析以下考试大纲，生成章节、小节、知识点、知识元的树状拓扑结构。

要求：
1. 输出格式为有效的JSON
2. 树状结构：章节(Chapter) -> 小节(Section) -> 知识点(KnowledgePoint) -> 知识元(KnowledgeElement)
3. 每个节点包含：name(名称)、description(详细内容/解释，50-200字)
4. 知识元是最底层的原子知识点，必须包含具体内容描述
5. 课程名称：{course_name}

请直接输出JSON，不要其他内容。JSON结构如下：
{{
    "course_name": "课程名称",
    "total_chapters": 章节数量,
    "chapters": [
        {{
            "name": "章节名称",
            "description": "章节内容描述",
            "sections": [
                {{
                    "name": "小节名称",
                    "description": "小节内容描述",
                    "knowledge_points": [
                        {{
                            "name": "知识点名称",
                            "description": "知识点的详细内容解释",
                            "elements": [
                                {{"name": "知识元名称", "description": "知识元的具体内容"}}
                            ]
                        }}
                    ]
                }}
            ]
        }}
    ]
}}

考试大纲内容：
{request.outline_text}"""

    def _parse_response(self, content: str) -> KnowledgeTree:
        data = json.loads(content)
        chapters = []
        for ch in data.get("chapters", []):
            sections = []
            for sec in ch.get("sections", []):
                knowledge_points = []
                for kp in sec.get("knowledge_points", []):
                    elements = []
                    for elem in kp.get("elements", []):
                        elements.append(
                            KnowledgeElement(
                                name=elem.get("name", ""),
                                description=elem.get("description"),
                            )
                        )
                    knowledge_points.append(
                        KnowledgePoint(
                            name=kp.get("name", ""),
                            description=kp.get("description"),
                            elements=elements,
                        )
                    )
                sections.append(
                    Section(
                        name=sec.get("name", ""),
                        description=sec.get("description"),
                        knowledge_points=knowledge_points,
                    )
                )
            chapters.append(
                Chapter(
                    name=ch.get("name", ""),
                    description=ch.get("description"),
                    sections=sections,
                )
            )

        return KnowledgeTree(
            course_name=data.get("course_name", ""),
            total_chapters=len(chapters),
            chapters=chapters,
        )

    async def analyze_outline(
        self, request: OutlineAnalysisRequest
    ) -> OutlineAnalysisResponse:
        try:
            prompt = self._build_prompt(request)
            content = await self._call_llm(prompt)
            knowledge_tree = self._parse_response(content)

            return OutlineAnalysisResponse(
                success=True,
                message="知识树生成成功",
                data=knowledge_tree,
            )
        except Exception as e:
            return OutlineAnalysisResponse(
                success=False,
                message=f"知识树生成失败: {str(e)}",
                data=None,
            )

    async def analyze_outline_stream(
        self, request: OutlineAnalysisRequest
    ) -> AsyncGenerator[str, None]:
        prompt = self._build_prompt(request)
        full_content = ""

        async for chunk in self._call_llm_stream(prompt):
            full_content += chunk
            yield f"data: {json.dumps({'type': 'chunk', 'content': chunk})}\n\n"

        try:
            knowledge_tree = self._parse_response(full_content)
            yield f"data: {json.dumps({'type': 'done', 'data': knowledge_tree.model_dump()})}\n\n"
        except Exception as e:
            yield f"data: {json.dumps({'type': 'error', 'message': str(e)})}\n\n"

    def save_to_database(self, db: Session, knowledge_tree: KnowledgeTree) -> str:
        course_name = knowledge_tree.course_name
        course_category = Category(
            name=course_name,
            parent_id=None,
            level=1,
            sort_order=0,
        )
        db.add(course_category)
        db.flush()

        chapter_sort = 0
        for chapter in knowledge_tree.chapters:
            chapter_category = Category(
                name=chapter.name,
                parent_id=course_category.id,
                level=2,
                sort_order=chapter_sort,
            )
            db.add(chapter_category)
            db.flush()

            section_sort = 0
            for section in chapter.sections:
                section_category = Category(
                    name=section.name,
                    parent_id=chapter_category.id,
                    level=3,
                    sort_order=section_sort,
                )
                db.add(section_category)
                db.flush()

                kp_sort = 0
                for kp in section.knowledge_points:
                    kp_category = Category(
                        name=kp.name,
                        parent_id=section_category.id,
                        level=4,
                        sort_order=kp_sort,
                    )
                    db.add(kp_category)
                    db.flush()

                    elem_sort = 0
                    for elem in kp.elements:
                        elem_category = Category(
                            name=elem.name,
                            parent_id=kp_category.id,
                            level=5,
                            sort_order=elem_sort,
                        )
                        db.add(elem_category)
                        db.flush()
                        elem_sort += 1
                    kp_sort += 1
                section_sort += 1
            chapter_sort += 1

        db.commit()
        return course_category.id


knowledge_tree_service = KnowledgeTreeService()
