# 智泉 · mixm

面向教学与在线实验的教育平台，将教学内容、作业、题库、模拟测试和容器实验环境集中到同一工作流中。

项目采用 Vue 3 前端与 Go 后端，支持通过 Kubernetes 分配实验实例，并可接入大语言模型。

[功能特性](#功能特性) · [快速开始](#快速开始) · [本地开发](#本地开发) · [配置说明](#配置说明) · [文档](#文档) · [参与贡献](#参与贡献)

## 功能特性

- **教学内容**：管理教程、文档、实验、视频与练习，使用分类树组织内容，并按角色控制可见范围。
- **作业管理**：发布作业、提交附件、批改评分与导出提交记录。
- **题库与测评**：题目分类、审核、练习组卷、模拟测试、答题记录与解析。
- **在线实验**：基于 Kubernetes 模板创建实例，提供 code-server、SSH 访问与持久化存储。
- **智能辅助**：接入大语言模型生成题目、分析知识树。
- **平台管理**：用户与角色权限、Excel 批量导入、镜像仓库管理和集群资源查看。

## 技术栈

| 模块 | 技术 |
| --- | --- |
| 前端 | Vue 3、Vite、Naive UI、Pinia、Vue Router |
| 后端 | Go 1.25.0、Gin、GORM、JWT |
| 数据与任务 | PostgreSQL、RabbitMQ、Redis |
| 实验环境 | Kubernetes、Client-go、code-server、SSH 网关 |
| 文件与镜像 | RustFS / S3 兼容存储、容器 Registry |

前端通过 `/api` 调用后端。后端将实例任务写入 PostgreSQL，并通过 RabbitMQ 交给独立 worker 执行，Redis 用于缓存任务结果。

## 快速开始

以下命令适用于 Bash，除特别说明外均在仓库根目录执行。首次构建需要能够拉取依赖包和容器镜像。

### 1. 获取代码并配置环境

准备 Git、Docker 和 Docker Compose，然后执行：

```bash
git clone https://github.com/Dragonzz27/mixm.git
cd mixm
cp backend/.env.example backend/.env
```

编辑 `backend/.env`，将 `SECRET_KEY` 替换为自己的随机密钥。初次体验无需准备 Kubernetes 集群。

默认数据库名、PostgreSQL 用户和 RabbitMQ 用户均为 `mixm`。如需调整，可在该文件中设置 `POSTGRES_DB`、`POSTGRES_USER`、`POSTGRES_PASSWORD`、`RABBITMQ_USER`、`RABBITMQ_PASSWORD` 以及 RustFS 凭据。数据库与 RabbitMQ 密码用于拼接连接 URL，需使用适合 URL 的值。

### 2. 启动应用

中间件由 `deploy/middleware/docker-compose.yaml` 统一管理，包含 PostgreSQL、RabbitMQ、Redis、RustFS 和 Registry。环境示例默认用于本地开发，Registry 绑定 `127.0.0.1:55000`，通过 `localhost:55000` 访问；跨主机部署时调整 `backend/.env` 中的 `REGISTRY_BIND_ADDRESS`，并将 `REGISTRY_URL`、`REGISTRY_PREFIX` 设置为相关主机与 Kubernetes 节点可访问的地址。

镜像默认从 Docker Hub 拉取。RustFS、Registry 和后端本地文件分别存放在项目根目录的 `.mixm/rustfs`、`.mixm/registry`、`.mixm/backend` 下。先以当前开发用户创建目录，再创建共享网络并启动中间件：

```bash
mkdir -p .mixm/{rustfs,registry,backend}

docker network inspect mixm-services >/dev/null 2>&1 || docker network create mixm-services

docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env up -d --wait
```

再启动前端与 API：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env up -d --build backend frontend
```

主 Compose 只包含 `backend`、`frontend` 和 `worker`。中间件项目与应用项目通过 `mixm-services` 网络和服务名互通；环境示例中的 `localhost` 地址用于本地 Go 开发。开发数据目录说明见[部署指南](deploy/README.md#持久化目录)。

### 3. 初始化数据与管理员

首次部署到新数据库时执行：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env \
  exec backend initdb

docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env exec \
  -e ADMIN_EMAIL=admin@example.com \
  -e ADMIN_PASSWORD='replace-with-your-password' \
  backend admin
```

将示例邮箱和密码替换为自己的值，然后使用该邮箱登录。`initdb` 初始化基础角色与内容类型，`admin` 仅创建新管理员，不会重置已有账号。学生账号可通过管理页面创建或批量导入。

### 4. 访问服务

| 服务 | 默认地址 |
| --- | --- |
| Web 界面 | <http://localhost:5173> |
| API 文档 | <http://localhost:8002/docs> |
| OpenAPI 描述 | <http://localhost:8002/openapi.json> |
| API 健康检查 | <http://localhost:8002/health> |
| RabbitMQ 管理界面 | <http://localhost:15672> |
| RustFS S3 接口 | <http://localhost:19000> |
| RustFS 控制台 | <http://localhost:19001> |

API 文档页面使用 CDN 加载 Swagger UI。未配置 Kubernetes 时，教学与用户功能可以运行，实例及集群相关接口返回 `503`。

查看日志或停止服务：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env logs -f backend
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env down
```

应用的 `down` 只停止前端、后端和 worker；中间件项目的管理命令见[部署指南](deploy/README.md)。

## 本地开发

准备 Go 1.25.0、Node.js 24（与前端镜像一致）和 pnpm，并按快速开始的第 1 步创建、编辑 `backend/.env`。

### 后端

先按快速开始第 2 步创建目录、网络并启动中间件。如已运行应用容器，先停止其中的 `backend`、`frontend` 和 `worker`，再启动 Go 服务：

```bash
cd backend
go mod download
go run ./cmd/initdb
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='replace-with-your-password' go run ./cmd/admin
go run ./cmd/server
```

如果已通过快速开始初始化数据库，可跳过 `initdb` 和 `admin`。

后端从当前工作目录读取 `.env`，因此 Go 命令需在 `backend/` 下执行。作业附件、导出文件、SSH 主机私钥统一保存在项目根目录的 `.mixm/backend`，后端会按需创建对应目录，具体路径见 `.env.example`。`.mixm` 已加入 Git 忽略规则。若修改了基础服务密码，请同步调整 `.env` 中的 `DATABASE_URL` 与 `RABBITMQ_URL`。

### 前端

在另一个终端从仓库根目录执行：

```bash
cd frontend
pnpm install
VITE_BACKEND_PROXY_TARGET=http://localhost:8002 pnpm dev
```

前端默认通过 `/api` 访问后端，上述命令将开发代理指向本机 Go 服务。也可以通过 `frontend/.env` 的 `VITE_API_BASE_URL` 指定 API 地址，格式参考 [前端环境示例](frontend/.env.example)。

## 可选功能

### Kubernetes 实验环境

1. 准备可用集群，将 kubeconfig 放到 `backend/kubeconfig/k3s.yaml`；容器部署时，配置中的集群地址必须可从容器内访问。
2. 配置 `K8S_SSH_GATEWAY_HOST`、`K8S_CODE_SERVER_HOST`、`STORAGE_CLASS_NAME`、`PVC_ACCESS_MODE` 与镜像仓库地址。
3. 应用新配置并启动 API 与 worker：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env up -d --build backend worker
```

本地开发时，在 `backend/` 下使用 `go run ./cmd/worker` 启动任务消费者；worker 需要有效的 Kubernetes 配置。SSH 网关默认端口为 `30022`。

存储类需要支持所配置的访问模式。实例模板与镜像部署方式见 [模板配置说明](backend/docs/template_config.md) 和 [部署指南](deploy/README.md)。

### 大语言模型

在 `backend/.env` 中设置 `LLM_API_URL`、`LLM_API_KEY` 和 `LLM_MODEL`，重启本地后端或重新创建后端容器后，可使用题目生成与知识树分析功能。

## 配置说明

完整配置见 [后端环境示例](backend/.env.example)。常用配置如下：

| 配置项 | 用途 |
| --- | --- |
| `SECRET_KEY` | JWT 签名密钥 |
| `DATABASE_URL`、`REDIS_URL`、`RABBITMQ_URL` | 本地开发的数据库、缓存与消息队列连接地址 |
| `DEPLOY_DATABASE_URL`、`DEPLOY_REDIS_URL`、`DEPLOY_RABBITMQ_URL`、`DEPLOY_RUSTFS_ENDPOINT`、`DEPLOY_REGISTRY_URL` | 覆盖应用容器的中间件连接地址，用于已有或远程服务 |
| `ACCESS_TOKEN_EXPIRE_MINUTES`、`REFRESH_TOKEN_EXPIRE_DAYS` | 访问令牌与刷新令牌有效期 |
| `K8S_CONFIG_PATH` | Kubernetes 配置文件路径 |
| `K8S_SSH_GATEWAY_HOST`、`K8S_CODE_SERVER_HOST` | 用户访问实验环境的主机地址 |
| `STORAGE_CLASS_NAME`、`PVC_ACCESS_MODE` | 实验持久化存储配置 |
| `REGISTRY_URL`、`REGISTRY_PREFIX` | 本地 Go 的镜像仓库 API 地址与 Kubernetes 可访问的镜像名称前缀 |
| `REGISTRY_BIND_ADDRESS`、`REGISTRY_PORT` | Registry 在宿主机上的监听地址与发布端口 |
| `RUSTFS_ENDPOINT`、`RUSTFS_BUCKET` | S3 兼容存储地址与桶名 |
| `LLM_API_URL`、`LLM_API_KEY`、`LLM_MODEL` | 大语言模型接口配置 |

应用 Compose 默认通过共享网络连接中间件项目；可用 `DEPLOY_*` 变量覆盖容器内的连接地址，具体值见 [Compose 配置](deploy/web/docker-compose.yaml)。环境文件和集群凭据应保留在本地。

## 测试与构建

后端：

```bash
cd backend
go test ./...
go vet ./...
make build
```

`make build` 在 `backend/bin/` 生成 `server`、`worker`、`admin` 和 `initdb` 四个程序。数据库结构通过嵌入二进制的编号 SQL 迁移自动初始化，基础角色和内容类型仍需通过 `initdb` 创建。

前端：

```bash
cd frontend
pnpm test
pnpm build
```

后端集成测试需显式提供独立测试服务；未设置对应变量时会跳过相关测试。先准备测试数据库 `mixm_test`，再从 `backend/` 执行：

```bash
TEST_DATABASE_URL='postgres://mixm:mixm@localhost:5432/mixm_test?sslmode=disable' \
TEST_RABBITMQ_URL='amqp://mixm:mixm@localhost:5672/' \
TEST_REDIS_URL='redis://localhost:6379/15' \
go test ./...
```

测试会创建并清理独立 PostgreSQL schema、RabbitMQ 队列和 Redis 缓存键。请使用测试环境，连接地址与凭据按实际服务调整。

## 项目结构

```text
.
├── backend/
│   ├── cmd/                # server、worker、admin、initdb 入口
│   ├── internal/
│   │   ├── app/            # 应用组装与启动
│   │   ├── router/         # 路由、OpenAPI 与接口契约
│   │   ├── handler/        # HTTP 请求与响应
│   │   ├── service/        # 业务逻辑与权限
│   │   ├── repository/     # 数据访问与 SQL 迁移
│   │   ├── model/          # 数据模型
│   │   ├── dto/            # 请求与响应类型
│   │   ├── auth/           # JWT 与密码处理
│   │   ├── middleware/     # HTTP 中间件
│   │   ├── config/         # 环境配置
│   │   └── infra/          # 外部客户端、集群、队列与 SSH
│   └── docs/               # API 与实验模板文档
├── frontend/               # Vue 应用与前端测试
├── deploy/                 # Compose、Dockerfile 与部署清单
└── scripts/                # 实验环境辅助脚本
```

## 文档

- [部署指南](deploy/README.md)：基础服务、镜像仓库、存储与 Kubernetes 部署。
- [REST API 说明](backend/docs/api.md)：响应格式、分页、资源接口与 SSE 协议。
- [路由清单](backend/internal/router/routes.json)：后端接口契约。
- [实验模板配置](backend/docs/template_config.md)：实例模板与环境配置。

## 参与贡献

欢迎通过 [Issues](https://github.com/Dragonzz27/mixm/issues) 报告问题或讨论需求，通过 [Pull Requests](https://github.com/Dragonzz27/mixm/pulls) 提交改进。

1. 提交问题时，说明运行环境、复现步骤、预期行为与实际结果；日志中请移除凭据和个人信息。
2. Fork 仓库并创建功能分支，尽量让每次改动聚焦一个问题。
3. 修改代码时补充必要的测试与文档；接口改动需同步前后端，数据库变更通过新增编号 SQL 迁移完成。
4. 运行相关测试与构建，在 Pull Request 中说明变更目的、验证方式及影响范围。

后端开发约定见 [backend/AGENTS.md](backend/AGENTS.md)。

## 许可证

仓库目前未提供 `LICENSE` 文件，具体授权范围待维护者明确。
