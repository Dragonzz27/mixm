# 部署文件

所有部署命令默认从项目根目录执行。

```text
deploy/
├── docker-compose.yaml                 # 仅 frontend、backend 和 worker
├── backend/
│   ├── Dockerfile                      # Go server、worker 和管理命令
│   └── Dockerfile.dockerignore         # 后端构建上下文排除规则
├── frontend/
│   ├── Dockerfile                      # Vue 开发服务镜像
│   └── Dockerfile.dockerignore         # 前端构建上下文排除规则
├── postgres/docker-compose.yaml        # 独立 PostgreSQL
├── rabbitmq/docker-compose.yaml        # 独立 RabbitMQ
├── redis/docker-compose.yaml           # 独立 Redis
├── registry/
│   ├── docker-compose.yaml            # 独立镜像仓库
│   └── ubuntu/                        # 实验镜像 Dockerfile 和 build.sh
├── rustfs/docker-compose.yaml          # 独立 RustFS 对象存储
└── longhorn/longhorn.yaml               # Longhorn 安装清单
```

## 启动顺序

应用与四个中间件各自使用独立的 Compose 项目。以下同机部署方式通过外部 Docker 网络 `mixm-services` 提供服务名解析，该网络只需创建一次。

先按[项目 README](../README.md)创建并编辑 `backend/.env`，再启动中间件：

```bash
docker network inspect mixm-services >/dev/null 2>&1 || docker network create mixm-services

docker compose -f deploy/postgres/docker-compose.yaml --env-file backend/.env up -d --wait
docker compose -f deploy/rabbitmq/docker-compose.yaml --env-file backend/.env up -d --wait
docker compose -f deploy/redis/docker-compose.yaml --env-file backend/.env up -d --wait
docker compose -f deploy/rustfs/docker-compose.yaml --env-file backend/.env up -d --wait
```

PostgreSQL、RabbitMQ 和 Redis 使用各自的健康检查；RustFS 的 `--wait` 等待容器进入运行状态。四个服务准备好后，启动应用：

```bash
docker compose -f deploy/docker-compose.yaml --env-file backend/.env up -d --build backend frontend
```

实验功能需要可用的 Kubernetes 配置；将 kubeconfig 放到 `backend/kubeconfig/k3s.yaml` 后，单独启动 worker：

```bash
docker compose -f deploy/docker-compose.yaml --env-file backend/.env up -d --build worker
```

主 Compose 不声明中间件，也不包含跨项目 `depends_on`；需先启动中间件，再启动应用。本地 Go 开发只需要上面的四份中间件配置。

升级时，后端启动会自动执行新增的 SQL 迁移。`003_remove_knowledge_tracing.sql` 删除已停用的掌握度与作答轨迹表，并清理历史测试结果中的掌握度和 Q 矩阵字段；试卷、答案、成绩与单次测试知识点正确率保留。

Dockerfile 位于此目录，构建上下文仍为原应用目录，因此 `COPY` 路径保持原样。单独构建镜像：

```bash
docker build -f deploy/backend/Dockerfile -t mixm-backend backend
docker build -f deploy/frontend/Dockerfile -t mixm-frontend frontend
```

前后端排除规则均采用与 Dockerfile 同目录的 `Dockerfile.dockerignore`，匹配的路径分别相对于 `backend/` 和 `frontend/` 构建上下文。前端会排除本地 `node_modules`、构建产物和环境文件，避免覆盖镜像内安装的依赖。

前端镜像默认监听 `0.0.0.0:5173`，与 Compose 配置一致；后端默认运行 `server`，同一镜像也提供 `worker`、`admin` 和 `initdb`。

Dockerfile 统一使用 `docker/dockerfile:1` 语法、大写指令、四空格续行缩进和 JSON 格式的 `CMD`。实验镜像使用 Bash 的 `pipefail`，脚本 heredoc 使用带引号的分隔符，让环境变量在容器运行时展开。

## 独立中间件管理

| 服务 | Compose 配置 | 项目名 | 数据位置 |
| --- | --- | --- | --- |
| PostgreSQL | `deploy/postgres/docker-compose.yaml` | `postgres` | 命名卷 `mixm-edu_postgres_data` |
| RabbitMQ | `deploy/rabbitmq/docker-compose.yaml` | `rabbitmq` | 命名卷 `mixm-edu_rabbitmq_data` |
| Redis | `deploy/redis/docker-compose.yaml` | `redis` | 命名卷 `mixm-edu_redis_data` |
| RustFS | `deploy/rustfs/docker-compose.yaml` | `rustfs` | 根目录 `RustFS/data` |
| Registry | `deploy/registry/docker-compose.yaml` | `registry` | 根目录 `registry/registry-data` |

例如单独查看、停止或重建 Redis：

```bash
docker compose -f deploy/redis/docker-compose.yaml --env-file backend/.env logs -f
docker compose -f deploy/redis/docker-compose.yaml --env-file backend/.env down
docker compose -f deploy/redis/docker-compose.yaml --env-file backend/.env up -d --wait
```

`deploy/docker-compose.yaml` 的 `down` 只管理前端、后端和 worker，不会停止独立中间件，也不会删除共享网络。中间件的 `down` 保留数据，`down -v` 会删除其管理的命名卷。

独立部署镜像仓库：

```bash
docker compose -f deploy/registry/docker-compose.yaml up -d
```

Registry 保留现有主机 IP 和端口绑定，可按部署主机修改其配置。它通过宿主机端口访问，不依赖 `mixm-services` 网络。

所有中间件配置都能分别启动；同机默认使用共享网络。应用和中间件统一使用 `--env-file backend/.env`，使 `POSTGRES_PASSWORD`、`RABBITMQ_PASSWORD` 和 RustFS 凭据一致。本地 Go 开发若修改密码，也需同步修改 `DATABASE_URL` 与 `RABBITMQ_URL`。

应用容器默认连接 `postgres:5432`、`rabbitmq:5672`、`redis:6379` 和 `http://rustfs:9000`。已有或远程服务可通过 `DEPLOY_DATABASE_URL`、`DEPLOY_RABBITMQ_URL`、`DEPLOY_REDIS_URL`、`DEPLOY_RUSTFS_ENDPOINT` 指定地址；这些变量只影响应用连接，不会修改独立中间件的配置。同机桥接网络不跨主机；PostgreSQL、RabbitMQ 和 Redis 默认仅发布到宿主机 `127.0.0.1`，跨主机部署时需要相应调整端口绑定。

### 已有部署切换

三个命名卷显式使用旧 `mixm-edu` 项目的卷名，RustFS 也保留原有目录。应用的 `backend_uploads`、`backend_static` 和 `backend_data` 仍属于 `mixm-edu`。如果历史部署使用了自定义项目名，应先把独立配置中的卷名改为实际已有的卷名。

RabbitMQ 数据与节点 hostname 关联，复用数据时需保留原节点名，见[官方迁移说明](https://www.rabbitmq.com/docs/backup)。切换前记录旧容器的 hostname（默认容器名示例如下，以实际名称为准），将输出写入 `backend/.env` 的 `RABBITMQ_HOSTNAME`：

```bash
docker inspect --format '{{.Config.Hostname}}' mixm-edu-rabbitmq-1
```

然后先停止旧的四个中间件容器，再按启动顺序启动独立配置。保留原数据卷与目录，不要同时运行两个使用同一数据卷的实例。新部署默认使用稳定的 `RABBITMQ_HOSTNAME=rabbitmq`。

## Kubernetes 与实验镜像

```bash
kubectl apply -f deploy/longhorn/longhorn.yaml
bash deploy/registry/ubuntu/build.sh
bash scripts/create-demo-pod.sh demo-user
kubectl apply -f scripts/code-server-demo-user.yaml
```

演示脚本位于根目录的 [`scripts/create-demo-pod.sh`](../scripts/create-demo-pod.sh)，生成的清单直接保存到 `scripts/`，文件名为 `code-server-<user_id>.yaml`；生成文件包含实例密码，已加入 Git 忽略规则。镜像构建脚本通过自身位置选择 Dockerfile，不依赖调用时的工作目录。

## 配置校验

```bash
docker compose -f deploy/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/postgres/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/rabbitmq/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/redis/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/rustfs/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/registry/docker-compose.yaml config --quiet
```

安装 Hadolint 后，可检查全部 7 个 Dockerfile；以下命令在发现语法或 Shell 错误时失败，版本固定等建议仍会显示：

```bash
hadolint --failure-threshold error \
  deploy/backend/Dockerfile \
  deploy/frontend/Dockerfile \
  deploy/registry/ubuntu/Dockerfile*
bash -n deploy/registry/ubuntu/build.sh
```

Docker daemon 可用时，还可以运行原生构建检查：

```bash
docker buildx build --check -f deploy/backend/Dockerfile backend
docker buildx build --check -f deploy/frontend/Dockerfile frontend
for dockerfile in deploy/registry/ubuntu/Dockerfile*; do
  docker buildx build --check -f "$dockerfile" deploy/registry/ubuntu
done
```
