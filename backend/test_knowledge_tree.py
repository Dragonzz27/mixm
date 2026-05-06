import asyncio
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from app.schema.knowledge_tree import OutlineAnalysisRequest
from app.service.knowledge_tree import knowledge_tree_service


async def main():
    test_outline = """
    第一章 线性代数基础
    1.1 行列式的定义与性质
    1.2 矩阵的运算
    1.3 逆矩阵

    第二章 向量空间
    2.1 向量的概念与运算
    2.2 线性相关性
    2.3 秩与极大线性无关组

    第三章 线性方程组
    3.1 高斯消元法
    3.2 解的结构
    """

    request = OutlineAnalysisRequest(
        outline_text=test_outline,
        course_name="线性代数",
    )

    result = await knowledge_tree_service.analyze_outline(request)
    print(f"Success: {result.success}")
    print(f"Message: {result.message}")
    if result.data:
        print(f"Course: {result.data.course_name}")
        print(f"Chapters: {result.data.total_chapters}")
        import json
        print(json.dumps(result.data.model_dump(), ensure_ascii=False, indent=2))


if __name__ == "__main__":
    asyncio.run(main())