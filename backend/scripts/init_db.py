#!/usr/bin/env python3
"""
数据库初始化脚本
初始化预设数据用于开发和测试
"""

import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from datetime import datetime
import argparse
from sqlalchemy.orm import Session
from app.database import engine, get_db, Base
from app.models.user import User
from app.models.template import Template
from app.models.category import Category
from app.models.content import Content
from app.models.content_type import ContentType
from app.models.instance import Instance
from app.models.role import Role
from app.models.user_role import UserRole


def init_users(db: Session) -> None:
    """初始化用户数据"""
    users = [
        User(
            username="admin",
            email="admin@mutim.com",
            hashed_password="$2b$12$xW2pZsIAcrBEt8zX1MgIxuBevXAe0N1/RUYQ31lC2A8taB3j6suba",
        ),
        User(
            username="testuser",
            email="test@mutim.com",
            hashed_password="$2b$12$xW2pZsIAcrBEt8zX1MgIxuBevXAe0N1/RUYQ31lC2A8taB3j6suba",
        ),
        User(
            username="developer",
            email="dev@mutim.com",
            hashed_password="$2b$12$xW2pZsIAcrBEt8zX1MgIxuBevXAe0N1/RUYQ31lC2A8taB3j6suba",
        ),
        User(
            username="operator",
            email="ops@mutim.com",
            hashed_password="$2b$12$xW2pZsIAcrBEt8zX1MgIxuBevXAe0N1/RUYQ31lC2A8taB3j6suba",
        ),
    ]

    for user in users:
        if not db.query(User).filter(User.username == user.username).first():
            db.add(user)

    db.commit()
    print(f"✓ 用户数据初始化完成 ({len(users)} 条)")


def init_roles(db: Session) -> None:
    """初始化角色数据"""
    roles = [
        {
            "name": "管理员",
            "route_keys": [
                "ClientHome",
                "Roles",
                "AdminImageRepositories",
                "ClientContents",
                "Categories",
                "Clusters",
                "ClientMyInstances",
                "Contents",
                "Dashboard",
                "Templates",
                "Users",
                "Instances",
            ],
        },
        {
            "name": "学生",
            "route_keys": [
                "ClientMyInstances",
                "ClientContents",
            ],
        },
    ]

    for role_data in roles:
        db_role = db.query(Role).filter(Role.name == role_data["name"]).first()
        if db_role:
            db_role.route_keys = role_data["route_keys"]
        else:
            db.add(Role(**role_data))

    db.commit()
    print(f"✓ 角色数据初始化完成 ({len(roles)} 条)")


def init_templates(db: Session) -> None:
    """初始化模板数据"""
    templates = [
        Template(
            name="后端API服务2",
            description="基于FastAPI的后端API服务模板，包含Python运行时和依赖配置",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="前端Web应用",
            description="Vue3前端应用模板，Nginx服务静态文件",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="数据库服务",
            description="PostgreSQL数据库服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="缓存服务",
            description="Redis缓存服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="Java后端服务",
            description="Spring Boot Java后端服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="消息队列",
            description="RabbitMQ消息队列服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="文件存储",
            description="MinIO对象存储服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
        Template(
            name="监控服务",
            description="Prometheus监控服务模板",
            config='{"image": "codercom/code-server:latest", "container_port": 8080, "cpu_limit": "1", "memory_limit": "1Gi", "cpu_request": "0.5", "memory_request": "512Mi", "storage": "1Gi", "working_dir": "/home/coder/project", "env": [{"name": "CS_DISABLE_PROXY_URI_VALIDATION", "value": "1"}, {"name": "CS_ALLOW_IFRAME", "value": "true"}]}',
        ),
    ]

    for template in templates:
        if not db.query(Template).filter(Template.name == template.name).first():
            db.add(template)

    db.commit()
    print(f"✓ 模板数据初始化完成 ({len(templates)} 条)")


def init_categories(db: Session) -> None:
    """初始化分类数据（五级：学科-课程-章节-小节-知识点）"""
    categories = [
        # L1 学科
        Category(
            name="计算机科学",
            description="计算机科学与技术学科",
            parent_id=None,
            level=1,
            sort_order=1,
        ),  # id=1
        Category(
            name="法律学",
            description="法学基础理论与应用",
            parent_id=None,
            level=1,
            sort_order=2,
        ),  # id=2
        # L2 课程 - 计算机科学
        Category(
            name="数据结构与算法",
            description="常用数据结构及算法分析",
            parent_id=1,
            level=2,
            sort_order=1,
        ),  # id=3
        Category(
            name="操作系统",
            description="操作系统原理与实践",
            parent_id=1,
            level=2,
            sort_order=2,
        ),  # id=4
        # L2 课程 - 法律学
        Category(
            name="民法",
            description="民事法律关系与权利",
            parent_id=2,
            level=2,
            sort_order=1,
        ),  # id=5
        Category(
            name="刑法",
            description="刑事犯罪与刑罚理论",
            parent_id=2,
            level=2,
            sort_order=2,
        ),  # id=6
        # L3 章节 - 数据结构与算法
        Category(
            name="线性结构",
            description="数组、链表等线性数据结构",
            parent_id=3,
            level=3,
            sort_order=1,
        ),  # id=7
        Category(
            name="树形结构",
            description="树、二叉树、图等非线性结构",
            parent_id=3,
            level=3,
            sort_order=2,
        ),  # id=8
        # L3 章节 - 操作系统
        Category(
            name="进程管理",
            description="进程调度与同步机制",
            parent_id=4,
            level=3,
            sort_order=1,
        ),  # id=9
        Category(
            name="内存管理",
            description="内存分配与虚拟内存",
            parent_id=4,
            level=3,
            sort_order=2,
        ),  # id=10
        # L3 章节 - 民法
        Category(
            name="民事权利",
            description="人格权、物权等民事权利体系",
            parent_id=5,
            level=3,
            sort_order=1,
        ),  # id=11
        Category(
            name="合同法",
            description="合同的成立、履行与违约",
            parent_id=5,
            level=3,
            sort_order=2,
        ),  # id=12
        # L3 章节 - 刑法
        Category(
            name="犯罪构成理论",
            description="犯罪的构成要件分析",
            parent_id=6,
            level=3,
            sort_order=1,
        ),  # id=13
        Category(
            name="刑罚体系",
            description="刑罚种类与量刑原则",
            parent_id=6,
            level=3,
            sort_order=2,
        ),  # id=14
        # L4 小节 - 线性结构
        Category(
            name="数组与字符串",
            description="数组操作及字符串处理",
            parent_id=7,
            level=4,
            sort_order=1,
        ),  # id=15
        Category(
            name="链表",
            description="单链表、双向链表与循环链表",
            parent_id=7,
            level=4,
            sort_order=2,
        ),  # id=16
        # L4 小节 - 树形结构
        Category(
            name="二叉树",
            description="二叉树的性质与遍历",
            parent_id=8,
            level=4,
            sort_order=1,
        ),  # id=17
        Category(
            name="图的基础",
            description="图的表示与基本遍历算法",
            parent_id=8,
            level=4,
            sort_order=2,
        ),  # id=18
        # L4 小节 - 进程管理
        Category(
            name="进程调度算法",
            description="常见调度算法原理与比较",
            parent_id=9,
            level=4,
            sort_order=1,
        ),  # id=19
        Category(
            name="进程同步与互斥",
            description="信号量、管程与死锁",
            parent_id=9,
            level=4,
            sort_order=2,
        ),  # id=20
        # L4 小节 - 民事权利
        Category(
            name="人格权",
            description="生命权、名誉权、隐私权等",
            parent_id=11,
            level=4,
            sort_order=1,
        ),  # id=21
        Category(
            name="物权",
            description="所有权、用益物权与担保物权",
            parent_id=11,
            level=4,
            sort_order=2,
        ),  # id=22
        # L4 小节 - 合同法
        Category(
            name="合同的成立",
            description="要约与承诺的规则",
            parent_id=12,
            level=4,
            sort_order=1,
        ),  # id=23
        Category(
            name="合同的履行与违约",
            description="履行规则及违约责任",
            parent_id=12,
            level=4,
            sort_order=2,
        ),  # id=24
        # L5 知识点 - 数组与字符串
        Category(
            name="数组的基本操作",
            description="增删查改、遍历与排序",
            parent_id=15,
            level=5,
            sort_order=1,
        ),  # id=25
        Category(
            name="多维数组与矩阵",
            description="二维数组与矩阵运算",
            parent_id=15,
            level=5,
            sort_order=2,
        ),  # id=26
        # L5 知识点 - 链表
        Category(
            name="单链表的增删查",
            description="单链表节点操作实现",
            parent_id=16,
            level=5,
            sort_order=1,
        ),  # id=27
        Category(
            name="双向链表与循环链表",
            description="双向链表结构与操作",
            parent_id=16,
            level=5,
            sort_order=2,
        ),  # id=28
        # L5 知识点 - 二叉树
        Category(
            name="二叉树的遍历",
            description="前序、中序、后序与层序遍历",
            parent_id=17,
            level=5,
            sort_order=1,
        ),  # id=29
        Category(
            name="二叉搜索树",
            description="BST的性质、插入与删除",
            parent_id=17,
            level=5,
            sort_order=2,
        ),  # id=30
        # L5 知识点 - 进程调度算法
        Category(
            name="先来先服务调度",
            description="FCFS算法原理与优缺点",
            parent_id=19,
            level=5,
            sort_order=1,
        ),  # id=31
        Category(
            name="优先级调度",
            description="静态与动态优先级调度策略",
            parent_id=19,
            level=5,
            sort_order=2,
        ),  # id=32
        # L5 知识点 - 人格权
        Category(
            name="生命权与健康权",
            description="生命权健康权的保护规则",
            parent_id=21,
            level=5,
            sort_order=1,
        ),  # id=33
        Category(
            name="名誉权与隐私权",
            description="名誉权隐私权的侵权认定",
            parent_id=21,
            level=5,
            sort_order=2,
        ),  # id=34
        # L5 知识点 - 合同的成立
        Category(
            name="要约与承诺",
            description="要约的构成要件及承诺规则",
            parent_id=23,
            level=5,
            sort_order=1,
        ),  # id=35
        Category(
            name="合同的形式要件",
            description="书面、口头及电子合同形式",
            parent_id=23,
            level=5,
            sort_order=2,
        ),  # id=36
    ]

    for category in categories:
        if (
            not db.query(Category)
            .filter(
                Category.name == category.name, Category.parent_id == category.parent_id
            )
            .first()
        ):
            db.add(category)

    db.commit()
    print(f"✓ 分类数据初始化完成 ({len(categories)} 条)")


def init_content_types(db: Session) -> None:
    """初始化内容类型数据"""
    content_types = [
        {"code": "tutorial", "name": "教程"},
        {"code": "documentation", "name": "文档"},
        {"code": "experiment", "name": "实验"},
        {"code": "video", "name": "视频"},
        {"code": "exercise", "name": "练习"},
    ]

    for ct in content_types:
        if not db.query(ContentType).filter(ContentType.code == ct["code"]).first():
            db.add(ContentType(**ct))

    db.commit()
    print(f"✓ 内容类型数据初始化完成 ({len(content_types)} 条)")


def init_contents(db: Session) -> None:
    """初始化内容数据"""
    contents = [
        # 计算机科学 - 数据结构与算法
        Content(
            title="数组与字符串入门教程",
            content="本教程介绍数组的基本概念、内存模型及常见操作，包括增删查改与边界处理...",
            content_type="tutorial",
            status="published",
            template_id=1,
            is_active=True,
        ),
        Content(
            title="链表原理与实现文档",
            content="详细说明单链表、双向链表的结构定义、节点操作及复杂度分析...",
            content_type="documentation",
            status="published",
            template_id=1,
            is_active=True,
        ),
        Content(
            title="链表反转实验",
            content="动手实现链表反转的多种方法（迭代法与递归法），并对比性能差异...",
            content_type="experiment",
            status="published",
            template_id=1,
            is_active=True,
        ),
        Content(
            title="二叉树遍历算法讲解视频",
            content="通过动画演示前序、中序、后序及层序四种遍历方式的执行过程...",
            content_type="video",
            status="published",
            template_id=2,
            is_active=True,
        ),
        Content(
            title="二叉搜索树专项练习",
            content="包含BST插入、删除、查找等10道经典题目，附参考答案与解析...",
            content_type="exercise",
            status="published",
            template_id=1,
            is_active=True,
        ),
        # 计算机科学 - 操作系统
        Content(
            title="进程调度算法教程",
            content="介绍FCFS、SJF、优先级调度及时间片轮转等常见调度策略的原理与比较...",
            content_type="tutorial",
            status="published",
            template_id=1,
            is_active=True,
        ),
        Content(
            title="进程同步与死锁实验",
            content="使用信号量模拟生产者-消费者问题，分析死锁产生条件及预防策略...",
            content_type="experiment",
            status="published",
            template_id=1,
            is_active=True,
        ),
        Content(
            title="操作系统内存管理文档",
            content="涵盖分页、分段、虚拟内存及页面置换算法的系统性参考文档...",
            content_type="documentation",
            status="draft",
            template_id=1,
            is_active=True,
        ),
        # 法律学 - 民法
        Content(
            title="民事权利体系概述教程",
            content="系统梳理人格权、物权、债权、知识产权等民事权利的分类与保护规则...",
            content_type="tutorial",
            status="published",
            template_id=2,
            is_active=True,
        ),
        Content(
            title="名誉权与隐私权侵权案例视频",
            content="结合典型判例，视频解析名誉权与隐私权在网络环境下的侵权认定要素...",
            content_type="video",
            status="published",
            template_id=2,
            is_active=True,
        ),
        # 法律学 - 合同法
        Content(
            title="要约与承诺规则文档",
            content="详解要约的构成要件、生效时间、撤回与撤销规则，以及承诺的形式要求...",
            content_type="documentation",
            status="published",
            template_id=2,
            is_active=True,
        ),
        Content(
            title="合同成立与违约责任练习",
            content="精选15道合同法综合练习题，覆盖合同成立、效力、履行及违约责任认定...",
            content_type="exercise",
            status="published",
            template_id=2,
            is_active=True,
        ),
    ]

    for content in contents:
        if not db.query(Content).filter(Content.title == content.title).first():
            db.add(content)

    db.commit()
    print(f"✓ 内容数据初始化完成 ({len(contents)} 条)")


def init_instances(db: Session) -> None:
    """初始化实例数据"""
    instances = [

    ]

    for instance in instances:
        if not db.query(Instance).filter(Instance.name == instance.name).first():
            db.add(instance)

    db.commit()
    print(f"✓ 实例数据初始化完成 ({len(instances)} 条)")


def main():
    """主函数"""
    parser = argparse.ArgumentParser(description="MutiM 数据库初始化")
    parser.add_argument("--force", "-f", action="store_true", help="Drop all tables before initializing")
    args = parser.parse_args()

    print("=" * 50)
    print("MutiM 数据库初始化")
    print("=" * 50)

    # 当传入 --force 时，直接删除所有表（drop_all）
    if args.force:
        print("⚠️ 已启用 --force：正在删除数据库中所有表（drop_all）...")
        Base.metadata.drop_all(bind=engine)
        print("✓ 所有表已删除")

    # 导入模型以保证 SQLAlchemy metadata 注册完整
    from app.models import user
    from app.models.category import Category as CategoryModel
    from app.models.content import Content as ContentModel
    from app.models.content_category import ContentCategory as ContentCategoryModel
    from app.models.content_type import ContentType as ContentTypeModel
    from app.models.template import Template as TemplateModel
    from app.models.instance import Instance as InstanceModel
    from app.models.role import Role as RoleModel
    from app.models.user_role import UserRole as UserRoleModel

    user.Base.metadata.create_all(bind=engine)
    CategoryModel.__table__.create(bind=engine, checkfirst=True)
    ContentModel.__table__.create(bind=engine, checkfirst=True)
    ContentCategoryModel.__table__.create(bind=engine, checkfirst=True)
    ContentTypeModel.__table__.create(bind=engine, checkfirst=True)
    TemplateModel.__table__.create(bind=engine, checkfirst=True)
    InstanceModel.__table__.create(bind=engine, checkfirst=True)
    RoleModel.__table__.create(bind=engine, checkfirst=True)
    UserRoleModel.__table__.create(bind=engine, checkfirst=True)
    # # 创建所有表（若已 drop_all 则为重新创建）
    # Base.metadata.create_all(bind=engine)

    print("✓ 数据库表创建完成\n")

    db = next(get_db())
    try:
        init_users(db)
        init_roles(db)
        init_templates(db)
        init_categories(db)
        init_content_types(db)
        init_contents(db)
        init_instances(db)

        print("\n" + "=" * 50)
        print("数据库初始化完成！")
        print("=" * 50)
    finally:
        db.close()


if __name__ == "__main__":
    main()
