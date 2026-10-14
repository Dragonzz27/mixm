> 本文保留历史功能设计说明。当前后端已迁移至 Go 1.25.0 + Client-go + RabbitMQ + PostgreSQL + Redis；运行与数据迁移请参阅 [README.md](README.md) 和 [迁移说明](backend/docs/migration.md)。

# 题库系统、考试系统、练习系统、作业系统设计方案

## 一、需求概述

在现有作业系统基础上，新增题库系统、考试系统、练习系统，统一进行题库管理。

### 1.1 核心功能

- **题库系统**：统一管理各类题目，支持LLM生成，带难度标识（1-5）
- **考试系统**：创建考试、发布考试（按角色）、记录成绩
- **练习系统**：创建练习、发布练习（按角色）、练习记录
- **作业系统**：现有系统优化改进（已有）

### 1.2 角色权限发布模式

参考现有作业系统的角色可见性控制：
- 创建时可指定可见角色列表
- 为空表示对所有角色可见
- 超级管理员可见所有

---

## 二、数据表设计

### 2.1 题库相关表

```sql
-- 题目表
CREATE TABLE questions (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,           -- 题目内容/题干
    type VARCHAR(20) NOT NULL,           -- 题型: single_choice/multi_choice/true_false/blank/essay
    content TEXT,                        -- 详细解析/内容
    answer TEXT,                         -- 答案
    explanation TEXT,                    -- 解析
    difficulty INTEGER NOT NULL,        -- 难度 1-5
    category_id VARCHAR(36),            -- 分类ID
    tags JSON,                         -- 标签列表
    source VARCHAR(100),               -- 来源
    is_ai_generated BOOLEAN DEFAULT FALSE,  -- 是否AI生成
    created_by VARCHAR(36),             -- 创建者ID
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- 题目分类表
CREATE TABLE question_categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,          -- 分类名称
    parent_id VARCHAR(36),              -- 父分类ID（支持多级）
    created_at TIMESTAMP DEFAULT NOW()
);

-- 题目-角色可见关联表
CREATE TABLE question_roles (
    question_id VARCHAR(36),
    role_id VARCHAR(36),
    PRIMARY KEY (question_id, role_id)
);

-- 题库选项表（选择题用）
CREATE TABLE question_options (
    id VARCHAR(36) PRIMARY KEY,
    question_id VARCHAR(36) NOT NULL,
    option_key VARCHAR(10) NOT NULL,    -- A/B/C/D
    content TEXT NOT NULL,            -- 选项内容
    is_correct BOOLEAN DEFAULT FALSE   -- 是否正确答案
);

-- 题目附件表
CREATE TABLE question_attachments (
    id VARCHAR(36) PRIMARY KEY,
    question_id VARCHAR(36) NOT NULL,
    file_path VARCHAR(500),
    file_name VARCHAR(200),
    file_type VARCHAR(50)
);
```

### 2.2 考试系统表

```sql
-- 考试表
CREATE TABLE exams (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,          -- 考试标题
    description TEXT,                    -- 考试描述
    start_time TIMESTAMP NOT NULL,       -- 开始时间
    end_time TIMESTAMP NOT NULL,         -- 结束时间
    duration_minutes INTEGER NOT NULL,   -- 考试时长（分钟）
    total_score INTEGER,                  -- 总分
    passing_score INTEGER,               -- 及格分
    status VARCHAR(20) DEFAULT 'draft',   -- 状态: draft/published/archived
    randomize BOOLEAN DEFAULT FALSE,      -- 是否随机选题
    show_score BOOLEAN DEFAULT TRUE,     -- 考试后是否显示成绩
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- 考试-题目关联表（组卷）
CREATE TABLE exam_questions (
    id VARCHAR(36) PRIMARY KEY,
    exam_id VARCHAR(36) NOT NULL,
    question_id VARCHAR(36) NOT NULL,
    score INTEGER DEFAULT 0,              -- 该题分值
    sort_order INTEGER DEFAULT 0           -- 排序
);

-- 考试-角色可见关联表
CREATE TABLE exam_roles (
    exam_id VARCHAR(36),
    role_id VARCHAR(36),
    PRIMARY KEY (exam_id, role_id)
);

-- 考试记录表
CREATE TABLE exam_records (
    id VARCHAR(36) PRIMARY KEY,
    exam_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    started_at TIMESTAMP,                   -- 开始作答时间
    submitted_at TIMESTAMP,               -- 提交时间
    score INTEGER,                      -- 得分
    status VARCHAR(20) DEFAULT 'not_started',  -- 状态: not_started/in_progress/submitted
    ip_address VARCHAR(50),
    PRIMARY KEY (exam_id, user_id)
);

-- 考试答题记录表
CREATE TABLE exam_answers (
    id VARCHAR(36) PRIMARY KEY,
    exam_record_id VARCHAR(36) NOT NULL,
    question_id VARCHAR(36) NOT NULL,
    answer TEXT,                         -- 用户答案
    score INTEGER,                      -- 该题得分
    is_correct BOOLEAN,
    answered_at TIMESTAMP
);
```

### 2.3 练习系统表

```sql
-- 练习表
CREATE TABLE practices (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,          -- 练习标题
    description TEXT,                    -- 练习描述
    category_id VARCHAR(36),             -- 练习分类
    difficulty_min INTEGER DEFAULT 1,     -- 最小难度
    difficulty_max INTEGER DEFAULT 5,     -- 最大难度
    question_count INTEGER,              -- 题目数量
    randomize BOOLEAN DEFAULT TRUE,     -- 是否随机选题
    show_answer BOOLEAN DEFAULT TRUE,  -- 提交后是否显示答案
    show_explanation BOOLEAN DEFAULT TRUE,  -- 是否显示解析
    status VARCHAR(20) DEFAULT 'draft',  -- 状态
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- 练习-题目关联表
CREATE TABLE practice_questions (
    id VARCHAR(36) PRIMARY KEY,
    practice_id VARCHAR(36) NOT NULL,
    question_id VARCHAR(36) NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- 练习-角色可见关联表
CREATE TABLE practice_roles (
    practice_id VARCHAR(36),
    role_id VARCHAR(36),
    PRIMARY KEY (practice_id, role_id)
);

-- 练习记录表
CREATE TABLE practice_records (
    id VARCHAR(36) PRIMARY KEY,
    practice_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    started_at TIMESTAMP,
    submitted_at TIMESTAMP,
    correct_count INTEGER,              -- 正确题数
    total_count INTEGER,               -- 总题数
    score INTEGER,                  -- 正确率
    PRIMARY KEY (practice_id, user_id)
);

-- 练习答题记录表
CREATE TABLE practice_answers (
    id VARCHAR(36) PRIMARY KEY,
    practice_record_id VARCHAR(36) NOT NULL,
    question_id VARCHAR(36) NOT NULL,
    user_answer TEXT,
    correct_answer TEXT,
    is_correct BOOLEAN,
    answered_at TIMESTAMP
);
```

### 2.4 作业系统改进（扩展现有表）

```sql
-- 可选：作业-题目关联表（作业也可以包含题目）
CREATE TABLE assignment_questions (
    id VARCHAR(36) PRIMARY KEY,
    assignment_id VARCHAR(36) NOT NULL,
    question_id VARCHAR(36) NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- 作业分类表
CREATE TABLE assignment_categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 三、API 设计

### 3.1 题库 API

```
题库管理（问题 CRUD）
- POST   /api/questions              创建题目
- GET    /api/questions             获取题目列表（分页、筛选）
- GET    /api/questions/{id}        获取题目详情
- PUT    /api/questions/{id}        更新题目
- DELETE /api/questions/{id}            删除题目
- POST   /api/questions/batch       批量操作

题库生成（LLM）
- POST   /api/questions/generate   LLM生成题目
- GET    /api/questions/ai-generated  获取AI生成的题目

题库分类
- POST   /api/question-categories
- GET    /api/question-categories
- PUT    /api/question-categories/{id}
- DELETE /api/question-categories/{id}
```

### 3.2 考试 API

```
考试管理
- POST   /api/exams                 创建考试
- GET    /api/exams                获取考试列表
- GET    /api/exams/{id}           获取考试详情
- PUT    /api/exams/{id}           更新考试
- DELETE /api/exams/{id}          删除考试
- POST   /api/exams/{id}/publish  发布考试
- POST   /api/exams/{id}/archive   归档考试

考试组卷
- POST   /api/exams/{id}/questions     添加题目
- DELETE /api/exams/{id}/questions    移除题目
- PUT    /api/exams/{id}/questions    调整题目顺序/分值

考试记录
- GET   /api/exams/{id}/records     查看考试记录
- GET   /api/exams/{id}/records/export  导出成绩
```

### 3.3 练习 API

```
练习管理
- POST   /api/practices            创建练习
- GET    /api/practices           获取练习列表
- GET    /api/practices/{id}      获取练习详情
- PUT    /api/practices/{id}      更新练习
- DELETE /api/practices/{id}     删除练习
- POST   /api/practices/{id}/publish  发布练习

练习记录
- GET   /api/practices/{id}/records    查看练习记录
- GET   /api/my/practices               我的练习记录
```

### 3.4 作业 API（扩展）

```
扩展现有作业API
- POST   /api/assignments          现有（创建作业）
- GET    /api/assignments          现有（获取列表）
- PUT    /api/assignments/{id}     现有
- DELETE /api/assignments/{id}     现有
- POST   /api/assignments/{id}/publish  发布作业（扩展）

作业分类
- GET    /api/assignment-categories   作业分类
- POST   /api/assignment-categories
```

### 3.5 统一发布 API

```
角色发布（所有系统统一模式）
- POST   /api/{resource_type}/{id}/roles      设置可见角色
- GET    /api/{resource_type}/{id}/roles      获取可见角色

{resource_type}: questions, exams, practices, assignments
```

---

## 四、LLM 生成题目设计

### 4.1 生成流程

```
LLM生成题目 -> 后端处理 -> 题目入库

请求参数：
- topic: string           主题/知识点
- difficulty: integer    目标难度 1-5
- count: integer         生成数量
- types: string[]        题型要求
- category_id: string    分类ID

生成方式：
1. 调用 LLM API 生成题目
2. 解析返回的题目数据
3. 批量创建题目记录
4. 标记 is_ai_generated = true
```

### 4.2 提示词示例

```
你是一个教育专家，请根据以下信息生成{count}道{difficulty}难度的{topic}题目。

要求：
1. 题型：{types}
2. 每道题包含：题干、答案、解析
3. 难度适中，适合{difficulty}级学习

请以JSON数组格式返回，每道题格式：
{{
  "title": "题干",
  "type": "题型",
  "answer": "答案",
  "explanation": "解析",
  "options": [{{"key": "A", "content": "选项内容"}}] // 选择题需要
}}
```

---

## 五、Service 层设计

### 5.1 服务类结构

```
app/
├── models/
│   ├── question.py           # 题目模型
│   ├── question_category.py  # 题目分类模型
│   ├── question_option.py    # 选项模型
│   ├── exam.py             # 考试模型
│   ├── exam_record.py      # 考试记录模型
│   ├── practice.py        # 练习模型
│   └── practice_record.py # 练习记录模型
│
├── schema/
│   ├── question.py         # 题目schema
│   ├── exam.py             # 考试schema
│   └── practice.py         # 练习schema
│
├── service/
│   ├── question.py         # 题目服务
│   ├── question_gen.py    # LLM生成服务
│   ├── exam.py           # ���试服务
│   └── practice.py      # 练习服务
│
└── router/
    ├── question.py       # 题目路由
    ├── exam.py          # 考试路由
    └── practice.py     # 练习路由
```

### 5.2 核心服务类

```python
# question_service.py
class QuestionService:
    def create(question: QuestionCreate, created_by: str) -> Question
    def get(question_id: str) -> Optional[Question]
    def get_list(skip, limit, filters) -> Tuple[List[Question], int
    def update(question_id: str, question: QuestionUpdate) -> Optional[Question]
    def delete(question_id: str) -> bool
    def batch_create(questions: List[QuestionCreate]) -> List[Question]

# question_gen_service.py
class QuestionGenService:
    def generate(topic, difficulty, count, types) -> List[Question]
    def parse_llm_response(response) -> List[dict]

# exam_service.py
class ExamService:
    def create(exam: ExamCreate, created_by: str) -> Exam
    def publish(exam_id: str, visible_roles: List[str]) -> bool
    def add_questions(exam_id: str, question_ids: List[str]) -> bool
    def start_exam(exam_id: str, user_id: str) -> ExamRecord
    def submit_exam(record_id: str, answers: List[Answer]) -> ExamRecord

# practice_service.py
class PracticeService:
    def create(practice: PracticeCreate, created_by: str) -> Practice
    def publish(practice_id: str, visible_roles: List[str]) -> bool
    def start_practice(practice_id: str, user_id: str) -> PracticeRecord
    def submit_practice(record_id: str, answers: List[Answer]) -> PracticeRecord
```

---

## 六、角色可见性实现

### 6.1 复用现有模式

参考 Assignment 的角色可见性实现：

```python
# 角色关联表
exam_roles = Table(
    "exam_roles",
    Base.metadata,
    Column("exam_id", String(36), ForeignKey("exams.id"), primary_key=True),
    Column("role_id", String(36), ForeignKey("roles.id"), primary_key=True),
)

class Exam(Base):
    __tablename__ = "exams"
    id = Column(String(36), primary_key=True)
    ...
    visible_roles = relationship("Role", secondary=exam_roles, back_populates="exams")
```

### 6.2 查询过滤逻辑

```python
def get_list_with_role_filter(db, user_roles, is_superuser):
    query = db.query(Exam)
    
    if not is_superuser:
        if user_roles:
            query = query.filter(
                or_(
                    ~Exam.id.in_(db.query(exam_roles.c.exam_id)),
                    Exam.id.in_(
                        db.query(exam_roles.c.exam_id).filter(
                            exam_roles.c.role_id.in_(user_roles)
                        )
                    ),
                )
            )
        else:
            query = query.filter(
                ~Exam.id.in_(db.query(exam_roles.c.exam_id))
            )
    
    return query.all()
```

---

## 七、实施步骤

### 7.1 第一阶段：题库系统

1. 创建数据表（questions, question_categories, question_options 等）
2. 实现 QuestionService
3. 实现 CRUD API
4. 实现题库分类管理

### 7.2 第二阶段：LLM 生成

1. 创建 QuestionGenService
2. 对接 LLM API
3. 实现批量生成逻辑
4. 处理题目解析和入库

### 7.3 第三阶段：考试系统

1. 创建考试相关表
2. 实现 ExamService
3. 实现考试发布（角色控制）
4. 实现考试作答和记录
5. 成绩导出

### 7.4 第四阶段：练习系统

1. 创建练习相关表
2. 实现 PracticeService
3. 实现练习发布（角色控制）
4. 实现练习作答和记录

### 7.5 第五阶段：作业系统优化

1. 添加作业分类
2. 扩展作业题目关联
3. 优化现有功能

---

## 八、文件结构

```
backend/app/
├── models/
│   ├── question.py           # 新增
│   ├── question_category.py # 新增
│   ├── question_option.py  # 新增
│   ├── exam.py              # 新增
│   ├── exam_record.py       # 新增
│   ├── practice.py        # 新增
│   └── practice_record.py # 新增
│
├── schema/
│   ├── question.py         # 新增
│   ├── exam.py           # 新增
│   └── practice.py      # 新增
│
├── service/
│   ├── question.py       # 新增
│   ├── question_gen.py # 新增
│   ├── exam.py        # 新增
│   └── practice.py   # 新增
│
└── router/
    ├── question.py      # 新增
    ├── exam.py        # 新增
    └── practice.py   # 新增
```

---

## 九、配置说明

### 9.1 环境变量

```bash
# LLM 配置
LLM_API_KEY=your_api_key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4

# 题目生成配置
QUESTION_GEN_DEFAULT_COUNT=10
QUESTION_GEN_DEFAULT_DIFFICULTY=3
```

### 9.2 数据库

现有使用 SQLite/PostgreSQL，新增表使用相同的数据库连接。