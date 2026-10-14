# -*- coding: utf-8 -*-
"""
初始化课程成绩和题库
初始化预设数据用于测试
"""


from app.database import SessionLocal
from app.models.category import Category
from app.models.question import Question, QuestionOption


def get_or_create_category(session, name, parent=None):
    parent_id = parent.id if parent else None
    existing = (
        session.query(Category)
        .filter(Category.name == name, Category.parent_id == parent_id)
        .first()
    )
    if existing:
        return existing

    category = Category(
        name=name,
        parent_id=parent_id,
        level=(parent.level + 1) if parent else 1,
        is_active=True,
    )
    session.add(category)
    session.flush()
    return category


def build_seed_specs():
    return [
        {
            "title": "以下关于数组和链表的说法，正确的是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "数组支持随机访问，链表插入删除更灵活",
            "explanation": "数组下标访问快，链表在中间插入删除通常更方便。",
            "knowledge_paths": [["计算机科学", "数据结构与算法"]],
            "options": [
                {"option_key": "A", "content": "数组插入删除一定比链表更快", "is_correct": False},
                {"option_key": "B", "content": "数组支持随机访问，链表插入删除更灵活", "is_correct": True},
                {"option_key": "C", "content": "链表一定比数组占用更少内存", "is_correct": False},
                {"option_key": "D", "content": "数组不能顺序访问", "is_correct": False},
            ],
        },
        {
            "title": "二叉搜索树的中序遍历结果一定是有序序列。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "二叉搜索树满足左子树 < 根 < 右子树，中序遍历天然有序。",
            "knowledge_paths": [["计算机科学", "数据结构与算法"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "下列哪些属于栈的典型应用？",
            "type": "multi_choice",
            "difficulty": 3,
            "answer": "表达式求值、括号匹配、函数调用",
            "explanation": "栈常用于需要后进先出的场景。",
            "knowledge_paths": [["计算机科学", "数据结构与算法"]],
            "options": [
                {"option_key": "A", "content": "表达式求值", "is_correct": True},
                {"option_key": "B", "content": "括号匹配", "is_correct": True},
                {"option_key": "C", "content": "函数调用", "is_correct": True},
                {"option_key": "D", "content": "按时间顺序排序", "is_correct": False},
            ],
        },
        {
            "title": "哈希表发生冲突时，常见的处理方法是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "拉链法或开放定址法",
            "explanation": "哈希冲突是常见问题，拉链法和开放定址法都很典型。",
            "knowledge_paths": [["计算机科学", "数据结构与算法"]],
            "options": [
                {"option_key": "A", "content": "拉链法或开放定址法", "is_correct": True},
                {"option_key": "B", "content": "二分查找法", "is_correct": False},
                {"option_key": "C", "content": "深度优先遍历", "is_correct": False},
                {"option_key": "D", "content": "冒泡排序", "is_correct": False},
            ],
        },
        {
            "title": "进程和线程的主要区别在于进程拥有独立的地址空间。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "线程共享进程资源，进程通常拥有独立地址空间。",
            "knowledge_paths": [["计算机科学", "操作系统"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "下列哪些属于进程同步机制？",
            "type": "multi_choice",
            "difficulty": 3,
            "answer": "信号量、互斥锁、条件变量",
            "explanation": "同步机制用于协调并发访问共享资源。",
            "knowledge_paths": [["计算机科学", "操作系统"]],
            "options": [
                {"option_key": "A", "content": "信号量", "is_correct": True},
                {"option_key": "B", "content": "互斥锁", "is_correct": True},
                {"option_key": "C", "content": "条件变量", "is_correct": True},
                {"option_key": "D", "content": "哈希表", "is_correct": False},
            ],
        },
        {
            "title": "LRU 页面置换算法的核心思想是优先淘汰________页面。",
            "type": "blank",
            "difficulty": 3,
            "answer": "最近最少使用",
            "explanation": "LRU 选择最近最久未被访问的页面进行淘汰。",
            "knowledge_paths": [["计算机科学", "操作系统"]],
        },
        {
            "title": "时间片轮转调度更适合哪类场景？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "多用户交互式系统",
            "explanation": "时间片轮转强调公平性和响应速度。",
            "knowledge_paths": [["计算机科学", "操作系统"]],
            "options": [
                {"option_key": "A", "content": "多用户交互式系统", "is_correct": True},
                {"option_key": "B", "content": "只运行单个批处理任务", "is_correct": False},
                {"option_key": "C", "content": "完全无需调度的系统", "is_correct": False},
                {"option_key": "D", "content": "只适合磁盘管理", "is_correct": False},
            ],
        },
        {
            "title": "SQL 中用于分组统计的关键字是？",
            "type": "single_choice",
            "difficulty": 1,
            "answer": "GROUP BY",
            "explanation": "GROUP BY 用于按字段分组后再聚合统计。",
            "knowledge_paths": [["计算机科学", "数据库"]],
            "options": [
                {"option_key": "A", "content": "ORDER BY", "is_correct": False},
                {"option_key": "B", "content": "GROUP BY", "is_correct": True},
                {"option_key": "C", "content": "WHERE", "is_correct": False},
                {"option_key": "D", "content": "LIMIT", "is_correct": False},
            ],
        },
        {
            "title": "事务的 ACID 特性包括哪些？",
            "type": "multi_choice",
            "difficulty": 3,
            "answer": "原子性、一致性、隔离性、持久性",
            "explanation": "ACID 是事务最重要的四个特性。",
            "knowledge_paths": [["计算机科学", "数据库"]],
            "options": [
                {"option_key": "A", "content": "原子性", "is_correct": True},
                {"option_key": "B", "content": "一致性", "is_correct": True},
                {"option_key": "C", "content": "隔离性", "is_correct": True},
                {"option_key": "D", "content": "持久性", "is_correct": True},
            ],
        },
        {
            "title": "在数据库设计中，第三范式的目标主要是减少________。",
            "type": "blank",
            "difficulty": 3,
            "answer": "数据冗余",
            "explanation": "3NF 通过消除传递依赖来减少数据冗余和更新异常。",
            "knowledge_paths": [["计算机科学", "数据库"]],
        },
        {
            "title": "索引的主要作用是提升查询性能。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "索引可以加快查询，但会增加写入和维护成本。",
            "knowledge_paths": [["计算机科学", "数据库"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "数组与链表的主要区别之一是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "数组支持随机访问，链表插入删除更灵活",
            "explanation": "数组随机访问快，链表插入删除更方便。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "线性结构", "数组与链表"]],
            "options": [
                {"option_key": "A", "content": "数组不支持随机访问", "is_correct": False},
                {"option_key": "B", "content": "数组支持随机访问，链表插入删除更灵活", "is_correct": True},
                {"option_key": "C", "content": "链表只能顺序存储", "is_correct": False},
                {"option_key": "D", "content": "链表一定比数组快", "is_correct": False},
            ],
        },
        {
            "title": "判断：栈是先进先出的线性结构。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "false",
            "explanation": "栈是后进先出（LIFO）。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "线性结构", "栈与队列"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": False},
                {"option_key": "B", "content": "错误", "is_correct": True},
            ],
        },
        {
            "title": "填空：串是由若干个________组成的有限序列。",
            "type": "blank",
            "difficulty": 1,
            "answer": "字符",
            "explanation": "串由字符构成。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "线性结构", "串与广义表"]],
        },
        {
            "title": "二叉树遍历常见顺序不包括以下哪一种？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "随机遍历",
            "explanation": "常见遍历有先序、中序、后序、层序。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "树与图", "二叉树遍历"]],
            "options": [
                {"option_key": "A", "content": "先序遍历", "is_correct": False},
                {"option_key": "B", "content": "中序遍历", "is_correct": False},
                {"option_key": "C", "content": "层序遍历", "is_correct": False},
                {"option_key": "D", "content": "随机遍历", "is_correct": True},
            ],
        },
        {
            "title": "最短路径算法中，Dijkstra 适用于哪类图？",
            "type": "single_choice",
            "difficulty": 3,
            "answer": "边权非负的图",
            "explanation": "Dijkstra 要求边权非负。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "树与图", "最短路径算法"]],
            "options": [
                {"option_key": "A", "content": "边权非负的图", "is_correct": True},
                {"option_key": "B", "content": "存在负权边的图", "is_correct": False},
                {"option_key": "C", "content": "有向无环图", "is_correct": False},
                {"option_key": "D", "content": "完全图", "is_correct": False},
            ],
        },
        {
            "title": "最小生成树的典型算法不包括以下哪一个？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "Dijkstra",
            "explanation": "Dijkstra 用于最短路径，不是最小生成树算法。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "树与图", "最小生成树"]],
            "options": [
                {"option_key": "A", "content": "Kruskal", "is_correct": False},
                {"option_key": "B", "content": "Prim", "is_correct": False},
                {"option_key": "C", "content": "Dijkstra", "is_correct": True},
                {"option_key": "D", "content": "Boruvka", "is_correct": False},
            ],
        },
        {
            "title": "快速排序的平均时间复杂度是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "O(n log n)",
            "explanation": "快速排序平均复杂度为 O(n log n)。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "排序与查找", "快速排序"]],
            "options": [
                {"option_key": "A", "content": "O(n)", "is_correct": False},
                {"option_key": "B", "content": "O(n log n)", "is_correct": True},
                {"option_key": "C", "content": "O(n^2)", "is_correct": False},
                {"option_key": "D", "content": "O(log n)", "is_correct": False},
            ],
        },
        {
            "title": "判断：二分查找要求待查序列是有序的。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "二分查找仅适用于有序序列。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "排序与查找", "二分查找"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "填空：哈希表冲突常用的解决方法是________。",
            "type": "blank",
            "difficulty": 2,
            "answer": "拉链法或开放定址法",
            "explanation": "哈希冲突常用拉链法或开放定址法处理。",
            "knowledge_paths": [["计算机科学", "数据结构与算法", "排序与查找", "哈希表"]],
        },
        {
            "title": "进程状态转换中，阻塞态转为就绪态通常是因为？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "等待的事件发生",
            "explanation": "阻塞进程等待的事件发生后进入就绪态。",
            "knowledge_paths": [["计算机科学", "操作系统", "进程管理", "进程状态转换"]],
            "options": [
                {"option_key": "A", "content": "时间片用完", "is_correct": False},
                {"option_key": "B", "content": "等待的事件发生", "is_correct": True},
                {"option_key": "C", "content": "CPU 空闲", "is_correct": False},
                {"option_key": "D", "content": "系统重启", "is_correct": False},
            ],
        },
        {
            "title": "调度算法中，能保证短作业优先的是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "最短作业优先",
            "explanation": "SJF 按作业长度排序。",
            "knowledge_paths": [["计算机科学", "操作系统", "进程管理", "调度算法"]],
            "options": [
                {"option_key": "A", "content": "时间片轮转", "is_correct": False},
                {"option_key": "B", "content": "最短作业优先", "is_correct": True},
                {"option_key": "C", "content": "先来先服务", "is_correct": False},
                {"option_key": "D", "content": "优先级调度", "is_correct": False},
            ],
        },
        {
            "title": "判断：互斥锁可用于实现临界区互斥访问。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "互斥锁用于保护临界区。",
            "knowledge_paths": [["计算机科学", "操作系统", "进程管理", "同步与互斥"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "填空：分页机制将内存划分为固定大小的________。",
            "type": "blank",
            "difficulty": 2,
            "answer": "页",
            "explanation": "分页把内存划分为固定大小的页。",
            "knowledge_paths": [["计算机科学", "操作系统", "内存管理", "分页与分段"]],
        },
        {
            "title": "虚拟内存的主要作用是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "扩展可用地址空间",
            "explanation": "虚拟内存扩展程序可用地址空间。",
            "knowledge_paths": [["计算机科学", "操作系统", "内存管理", "虚拟内存"]],
            "options": [
                {"option_key": "A", "content": "提高 CPU 主频", "is_correct": False},
                {"option_key": "B", "content": "扩展可用地址空间", "is_correct": True},
                {"option_key": "C", "content": "禁止进程切换", "is_correct": False},
                {"option_key": "D", "content": "减少线程数量", "is_correct": False},
            ],
        },
        {
            "title": "页面置换算法中，FIFO 的含义是？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "先进先出",
            "explanation": "FIFO 选择最早进入内存的页面置换。",
            "knowledge_paths": [["计算机科学", "操作系统", "内存管理", "页面置换算法"]],
            "options": [
                {"option_key": "A", "content": "最少使用", "is_correct": False},
                {"option_key": "B", "content": "先进先出", "is_correct": True},
                {"option_key": "C", "content": "最久未用", "is_correct": False},
                {"option_key": "D", "content": "随机替换", "is_correct": False},
            ],
        },
        {
            "title": "判断：目录结构用于组织文件并支持层次化管理。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "目录结构用于管理文件层次关系。",
            "knowledge_paths": [["计算机科学", "操作系统", "文件系统", "目录结构"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "磁盘调度算法中，SCAN 又称为？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "电梯算法",
            "explanation": "SCAN 常被称为电梯算法。",
            "knowledge_paths": [["计算机科学", "操作系统", "文件系统", "磁盘调度"]],
            "options": [
                {"option_key": "A", "content": "电梯算法", "is_correct": True},
                {"option_key": "B", "content": "最短寻道优先", "is_correct": False},
                {"option_key": "C", "content": "先来先服务", "is_correct": False},
                {"option_key": "D", "content": "循环调度", "is_correct": False},
            ],
        },
        {
            "title": "填空：FAT 文件系统使用的文件分配结构是________表。",
            "type": "blank",
            "difficulty": 2,
            "answer": "文件分配",
            "explanation": "FAT 即文件分配表。",
            "knowledge_paths": [["计算机科学", "操作系统", "文件系统", "文件分配表"]],
        },
        {
            "title": "判断：光纤属于有线传输介质。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "光纤是典型的有线传输介质。",
            "knowledge_paths": [["计算机科学", "计算机网络", "物理层", "传输介质"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "以下哪一项属于数字调制技术？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "QAM",
            "explanation": "QAM 是常见的数字调制方式。",
            "knowledge_paths": [["计算机科学", "计算机网络", "物理层", "编码与调制"]],
            "options": [
                {"option_key": "A", "content": "QAM", "is_correct": True},
                {"option_key": "B", "content": "HTTP", "is_correct": False},
                {"option_key": "C", "content": "IP", "is_correct": False},
                {"option_key": "D", "content": "DNS", "is_correct": False},
            ],
        },
        {
            "title": "填空：在频分复用中，不同信号占用不同的________。",
            "type": "blank",
            "difficulty": 2,
            "answer": "频率带宽",
            "explanation": "FDM 通过不同频率带宽复用信道。",
            "knowledge_paths": [["计算机科学", "计算机网络", "物理层", "复用技术"]],
        },
        {
            "title": "判断：IP 协议属于网络层协议。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "IP 位于网络层，负责寻址与转发。",
            "knowledge_paths": [["计算机科学", "计算机网络", "网络层", "IP协议"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "路由算法中，最短路径路由属于哪一类？",
            "type": "single_choice",
            "difficulty": 2,
            "answer": "链路状态路由",
            "explanation": "链路状态路由常用最短路径算法。",
            "knowledge_paths": [["计算机科学", "计算机网络", "网络层", "路由算法"]],
            "options": [
                {"option_key": "A", "content": "距离向量路由", "is_correct": False},
                {"option_key": "B", "content": "链路状态路由", "is_correct": True},
                {"option_key": "C", "content": "应用层路由", "is_correct": False},
                {"option_key": "D", "content": "物理层路由", "is_correct": False},
            ],
        },
        {
            "title": "填空：子网划分的主要目的是减少________和提升管理效率。",
            "type": "blank",
            "difficulty": 2,
            "answer": "广播域",
            "explanation": "划分子网可以缩小广播域。",
            "knowledge_paths": [["计算机科学", "计算机网络", "网络层", "子网划分"]],
        },
        {
            "title": "HTTP 属于哪一层协议？",
            "type": "single_choice",
            "difficulty": 1,
            "answer": "应用层",
            "explanation": "HTTP 是应用层协议。",
            "knowledge_paths": [["计算机科学", "计算机网络", "应用层", "HTTP协议"]],
            "options": [
                {"option_key": "A", "content": "应用层", "is_correct": True},
                {"option_key": "B", "content": "传输层", "is_correct": False},
                {"option_key": "C", "content": "网络层", "is_correct": False},
                {"option_key": "D", "content": "数据链路层", "is_correct": False},
            ],
        },
        {
            "title": "判断：DNS 主要用于将域名解析为 IP 地址。",
            "type": "true_false",
            "difficulty": 1,
            "answer": "true",
            "explanation": "DNS 提供域名到 IP 的解析服务。",
            "knowledge_paths": [["计算机科学", "计算机网络", "应用层", "DNS解析"]],
            "options": [
                {"option_key": "A", "content": "正确", "is_correct": True},
                {"option_key": "B", "content": "错误", "is_correct": False},
            ],
        },
        {
            "title": "以下哪一项属于电子邮件协议？",
            "type": "multi_choice",
            "difficulty": 2,
            "answer": "SMTP、POP3",
            "explanation": "常见邮件协议包括 SMTP、POP3、IMAP。",
            "knowledge_paths": [["计算机科学", "计算机网络", "应用层", "电子邮件协议"]],
            "options": [
                {"option_key": "A", "content": "SMTP", "is_correct": True},
                {"option_key": "B", "content": "POP3", "is_correct": True},
                {"option_key": "C", "content": "FTP", "is_correct": False},
                {"option_key": "D", "content": "ARP", "is_correct": False},
            ],
        },
    ]


def build_course_hierarchy():
    return {
        "计算机科学": {
            "数据结构与算法": {
                "线性结构": ["数组与链表", "栈与队列", "串与广义表"],
                "树与图": ["二叉树遍历", "最短路径算法", "最小生成树"],
                "排序与查找": ["快速排序", "二分查找", "哈希表"],
            },
            "操作系统": {
                "进程管理": ["进程状态转换", "调度算法", "同步与互斥"],
                "内存管理": ["分页与分段", "虚拟内存", "页面置换算法"],
                "文件系统": ["目录结构", "磁盘调度", "文件分配表"],
            },
            "数据库系统": {
                "关系模型": ["关系代数", "完整性约束", "范式理论"],
                "SQL语言": ["数据查询", "数据更新", "视图与索引"],
                "事务管理": ["ACID特性", "并发控制", "故障恢复"],
            },
            "计算机网络": {
                "物理层": ["传输介质", "编码与调制", "复用技术"],
                "网络层": ["IP协议", "路由算法", "子网划分"],
                "应用层": ["HTTP协议", "DNS解析", "电子邮件协议"],
            },
            "编程基础": {
                "语法与结构": ["变量与数据类型", "控制流语句", "数组与字符串"],
                "函数与模块": ["参数传递", "作用域规则", "递归函数"],
                "面向对象": ["类与对象", "继承与多态", "封装与接口"],
            },
        }
    }


def seed_course_categories(session):
    hierarchy = build_course_hierarchy()
    for course_name, chapters in hierarchy.items():
        course = get_or_create_category(session, course_name)
        for chapter_name, sections in chapters.items():
            chapter = get_or_create_category(session, chapter_name, parent=course)
            for section_name, topics in sections.items():
                section = get_or_create_category(session, section_name, parent=chapter)
                for topic_name in topics:
                    get_or_create_category(session, topic_name, parent=section)


def add_or_update_question(session, spec, existing=None):
    question = existing or Question()
    question.title = spec["title"]
    question.type = spec["type"]
    question.difficulty = spec["difficulty"]
    question.answer = spec.get("answer")
    question.explanation = spec.get("explanation")
    question.review_status = "approved"
    question.is_ai_generated = False
    question.source = "seed_questions"

    if existing is None:
        session.add(question)
        session.flush()

    question.knowledge_tags = []
    if spec.get("knowledge_paths"):
        tags = []
        for path in spec["knowledge_paths"]:
            parent = None
            for name in path:
                parent = get_or_create_category(session, name, parent=parent)
            if parent:
                tags.append(parent)
        question.knowledge_tags = tags

    session.query(QuestionOption).filter(
        QuestionOption.question_id == question.id
    ).delete(synchronize_session=False)

    for opt in spec.get("options") or []:
        session.add(
            QuestionOption(
                question_id=question.id,
                option_key=opt["option_key"],
                content=opt["content"],
                is_correct=opt.get("is_correct", False),
            )
        )

    return question


def seed_questions(session):
    seed_course_categories(session)
    specs = build_seed_specs()

    existing_seed_questions = (
        session.query(Question)
        .filter(Question.source == "seed_questions")
        .order_by(Question.created_at.asc().nullslast(), Question.id.asc())
        .all()
    )

    for index, spec in enumerate(specs):
        existing = existing_seed_questions[index] if index < len(existing_seed_questions) else None
        add_or_update_question(session, spec, existing=existing)


def main():
    session = SessionLocal()
    try:
        seed_questions(session)
        session.commit()
        count = session.query(Question).count()
        print(f"Seed completed. Total questions: {count}")
    finally:
        session.close()


if __name__ == "__main__":
    main()
