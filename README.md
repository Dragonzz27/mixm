# 智泉 / MutiM

教育平台，前端使用 Vue，后端使用 **Go 1.25.0 + Gin + GORM + Client-go + RabbitMQ + PostgreSQL + Redis**。

包含用户与角色管理、教学内容与知识树、作业提交及评分、题库审核、练习卷、模拟测试、Kubernetes 实验实例、code-server 和 SSH 网关。RustFS 提供 S3 文件存储，Registry 提供容器镜像。`kt_service` 是独立的 Python/PyTorch 模型推理服务，由 Go 通过 HTTP 调用。

## Docker Compose

部署文件统一位于 [`deploy/`](deploy/README.md)，下列部署命令均从项目根目录执行。

```bash
cp backend/.env.example backend/.env
mkdir -p backend/kubeconfig
# 将可从容器访问的 kubeconfig 放到 backend/kubeconfig/k3s.yaml
# 修改 backend/.env 中的 SECRET_KEY、Kubernetes 主机、Registry 和存储类配置

docker compose -f deploy/docker-compose.yaml --env-file backend/.env up -d --build
```

Compose 的 PostgreSQL、RabbitMQ、Redis 地址自动使用服务名；`.env.example` 中的 localhost 地址用于本地 Go 开发。`POSTGRES_PASSWORD`、`RABBITMQ_PASSWORD` 可在同一环境文件中覆盖 Compose 默认值，连接字符串中的密码需适合 URL 使用。

首次创建新数据库后初始化角色、内容类型和管理员：

```bash
docker compose -f deploy/docker-compose.yaml --env-file backend/.env exec backend initdb
docker compose -f deploy/docker-compose.yaml --env-file backend/.env exec \
  -e ADMIN_EMAIL=admin@example.com \
  -e ADMIN_PASSWORD='replace-with-your-password' \
  backend admin
```

管理员命令只创建新账号，不会重置已有账号；学生账号由管理页面创建或导入。Excel 导入使用前端模板的 `username,email,phone,role_name` 列和默认密码 `123456`，也可提供 `password`、`role_id` 列。

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:5173 |
| API | http://localhost:8002/api |
| API 文档 | http://localhost:8002/docs |
| OpenAPI | http://localhost:8002/openapi.json |
| 健康检查 | http://localhost:8002/health |
| SSH 网关 | localhost:30022 |
| RabbitMQ 管理界面 | http://localhost:15672 |
| RustFS | http://localhost:19000 |

API 文档使用 Swagger UI；页面资源需要访问 CDN。无 Kubernetes 配置时，HTTP 服务可运行教学和用户功能，实例及集群接口返回 503，worker 需要有效 kubeconfig。

## 本地开发

使用当前 `go version go1.25.0 linux/amd64`，无需更换 Go 工具链：

```bash
docker compose -f deploy/docker-compose.yaml up -d postgres rabbitmq redis
cd backend
cp .env.example .env
# 编辑 .env，设置 SECRET_KEY 和 kubeconfig 等参数
go mod download
go run ./cmd/initdb
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='replace-with-your-password' go run ./cmd/admin
go run ./cmd/server
```

在另一终端从 `backend/` 启动任务消费者：

```bash
go run ./cmd/worker
```

前端（从项目根目录执行）：

```bash
cd frontend
pnpm install
pnpm dev
```

本地前端开发需将 `frontend/vite.config.js` 的代理目标设为 `http://localhost:8002`，或设置 `VITE_BACKEND_PROXY_TARGET=http://localhost:8002`；Compose 使用默认目标 `http://backend:8002`。

## 数据库初始化

应用使用 PostgreSQL，server、worker、admin 和 initdb 启动时自动执行数据库结构初始化。SQL 文件嵌入二进制，schema_migrations 记录已执行的版本，事务锁防止多个进程重复建表。

首次启动后执行 initdb 创建基础角色和内容类型，再用 admin 创建管理员。

## 验证与构建

```bash
cd backend
go test ./...
go vet ./...
make build
```

集成测试只在指定测试服务时运行，自动创建和清理独立 PostgreSQL schema、RabbitMQ 队列、Redis 缓存键：

```bash
TEST_DATABASE_URL='postgres://mutim:mutim@localhost:5432/mutim_test?sslmode=disable' \
TEST_RABBITMQ_URL='amqp://mutim:mutim@localhost:5672/' \
TEST_REDIS_URL='redis://localhost:6379/15' \
go test ./...
```

`TestAPIRouteCoverage` 检查当前 134 条 API 接口。其他测试覆盖 JWT 边界、权限、关联数据、服务端评分、试卷恢复、Kubernetes 资源幂等、RabbitMQ 重复投递和 Redis 结果缓存。Kubernetes 测试使用 Client-go fake client。

## 代码结构

```text
backend/
├── cmd/                   # server、worker、admin、initdb
├── internal/router/       # Gin 路由、OpenAPI 与接口契约测试
├── internal/middleware/   # JWT 身份认证、CORS、日志、异常恢复
├── internal/handler/      # 参数绑定及 JSON、文件、SSE 输出
├── internal/dto/          # 按业务拆分的请求 DTO、可选字段、响应包裹与分页
├── internal/service/      # 权限、业务规则、事务协调
├── internal/repository/   # GORM 连接、编号 SQL 迁移、查询、事务与任务连接锁
├── internal/model/        # 现有 PostgreSQL 表映射和动态业务投影
├── internal/app/          # 依赖注入、四个入口的启动与关闭
├── internal/auth/         # JWT、bcrypt
├── internal/apperror/     # 客户端、业务层与 HTTP 层共用的应用错误
├── internal/config/       # 环境配置
├── internal/infra/        # 基础设施目录分组，各子目录仍是独立 Go 包
│   ├── client/            # Registry、S3、LLM、KT 外部客户端
│   ├── cluster/           # Client-go 资源生命周期
│   ├── queue/             # PostgreSQL outbox、RabbitMQ publisher/worker、Redis 缓存
│   └── ssh/               # 通过 repository 查询实例的 SSH exec 网关
├── internal/testutil/     # 独立测试数据库等项目专用测试支持
├── docs/                  # 实例模板配置说明
└── kubeconfig/            # 本地集群凭据，不包含在镜像内
deploy/                    # Compose、Dockerfile、Registry 镜像、Kubernetes 清单与脚本
scripts/                   # 通用辅助脚本，例如 create-demo-pod.sh
frontend/                  # Vue 前端
kt_service/                # 独立模型推理服务
```

项目为专有项目，保密使用。

四个命令入口共享 `internal` 包；辅助函数留在其所属的职责包内，不集中到统一的 `utils` 包。当前没有跨项目 Go 复用需求，因此不设置 `backend/pkg`。应用错误使用 `apperror.Error`，请求与响应类型放在 `dto`，数据库连接和迁移统一由 `repository` 管理。编号 SQL 位于 `internal/repository/migrations`，文件名、SQL 内容与迁移记录保持兼容。

## REST API

后端与前端需要同步更新部署。业务接口仍位于 `/api`，普通 JSON 使用 `{"code":0,"message":"成功","data":...}`；失败保留 HTTP 错误状态，将说明放在 `message`，`data` 为 `null`。列表统一为 `data.items/total/offset/limit`，默认 `offset=0`、`limit=100`，部分更新使用 `PATCH`。

旧业务接口已移除；历史附件 `/api/assignments/file/`、`/api/contents/file/` 和 `/static/` 链接保留。SSE、文件和基础设施接口使用各自格式。完整路由清单见 [routes.json](backend/internal/router/routes.json)，行为与改版映射见 [API 文档](backend/docs/api.md)。

数据库继续使用已有编号 SQL 迁移，不执行 GORM AutoMigrate，也不重写已应用迁移。worker 在同一连接上持有 PostgreSQL advisory lock，保留任务重投递和 outbox 语义。

前端协议及 SSE 解析测试：`cd frontend && npm test`。后端测试仍使用独立 PostgreSQL schema、RabbitMQ 队列和 Redis 测试服务。
