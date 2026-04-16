# 智泉

## TODO
现在在HyperV中进行的部署k3s，但是code-server在目前iframe里可以打开，但输入密码后还是重定向到密码框。cd

## 部署

### backend

1. k3s部署后，提取kubeconfig文件，保存至/backend/app/kubeconfig/k3s.yaml
2. 检查.env中是否为 K8S_CONFIG_PATH=kubeconfig/k3s.yaml
3. 注意cluster.cluster.server 非127.0.0.1

部署docker-compose.yaml采用nerdctl，注意postgresql文件夹需赋予权限


### fronted

注意配置.env



# MutiM

## 项目简介

MutiM 是一个全栈应用，后端使用 FastAPI（Python），前端使用 Vue.js。

## 功能概览

- 用户认证与授权
- 基于角色的作业可见性控制
- 作业管理（创建、更新、删除）
- 提交管理（上传、审核、评分）
- 将作业成绩导出为 CSV

## 快速开始

### 前置条件

- 推荐使用 Docker 与 Docker Compose
- 或本地开发环境：
  - Python 3.11+
  - Node.js 20+

### 使用 Docker Compose 启动（开发推荐）

1. 克隆仓库
2. 在项目根目录运行：
   ```bash
   docker-compose up -d
   ```
3. 启动后服务地址：
   - 前端: http://localhost:3000
   - 后端 API: http://localhost:8002/api
   - API 文档: http://localhost:8002/docs

### 本地开发

#### 后端

1. 进入后端目录：
   ```bash
   cd backend
   ```
2. 创建并激活虚拟环境：
   ```bash
   python -m venv venv
   # macOS/Linux: source venv/bin/activate
   # Windows: venv\\Scripts\\activate
   ```
3. 安装依赖：
   ```bash
   pip install -r requirements.txt
   ```
4. 复制并配置环境变量（将 `.env.example` 另存为 `.env` 并调整）
5. 运行开发服务器：
   ```bash
   uvicorn app.main:app --host 0.0.0.0 --port 8002 --reload
   ```

#### 前端

1. 进入前端目录：
   ```bash
   cd frontend
   ```
2. 安装依赖：
   ```bash
   npm install
   ```
3. 配置环境变量（根据 `.env.example` 创建 `.env`）
4. 启动开发服务器：
   ```bash
   npm run dev
   ```
5. 前端默认地址： http://localhost:5173

## API 文档

启动后访问 http://localhost:8002/docs 可查看交互式 API 文档。

## 详细功能说明

### 基于角色的作业可见性

- 管理员在创建/编辑作业时可指定可见角色
- 若未选择任何角色，则该作业对所有用户可见
- 超级管理员（`is_superuser=true`）可查看所有作业
- 普通用户仅能看到：
  - 对所有角色可见的作业（未选择角色）
  - 与其拥有的角色至少有交集的作业

### 提交管理

- 用户可上传文件提交作业（支持 doc、docx、pdf、ppt、pptx、txt）
- 教师可查看提交、给出反馈并评分
- 提交状态：submitted、reviewed、completed

### 成绩导出

- 教师可将某作业的提交成绩导出为 CSV 文件
- CSV 包含：提交 ID、用户 ID、用户名、文件名、提交时间、状态、成绩、反馈

## 配置说明

### 后端（.env）

- `SECRET_KEY`：用于 JWT 签发的密钥
- `ALGORITHM`：JWT 算法（默认：HS256）
- `ACCESS_TOKEN_EXPIRE_MINUTES`：访问令牌过期时间（分钟）
- `DATABASE_URL`：数据库连接字符串（默认：SQLite）

### 前端

- `VITE_API_BASE_URL`：API 基础地址（默认：http://localhost:8002/api）

## 项目结构

```
mixm-edu/
├── backend/
│   ├── app/                 # FastAPI 应用代码
   │   ├── models/           # SQLAlchemy 模型
   │   ├── router/           # 路由定义
   │   ├── schema/           # Pydantic 模型
   │   └── service/          # 业务逻辑
├── frontend/                # Vue 前端
├── docker-compose.yaml      # Docker Compose 配置
└── README.md                # 项目说明（此文件）
```

## 许可证

本项目为专有项目，保密使用。
