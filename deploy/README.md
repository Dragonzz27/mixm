# 部署文件

所有部署命令默认从项目根目录执行。

```text
deploy/
├── web/
│   ├── docker-compose.yaml             # frontend、backend 和 worker
│   ├── Dockerfile.backend              # Go server、worker 和管理命令
│   ├── Dockerfile.backend.dockerignore  # 后端构建上下文排除规则
│   ├── Dockerfile.frontend             # Vue 开发服务镜像
│   └── Dockerfile.frontend.dockerignore # 前端构建上下文排除规则
├── middleware/
│   └── docker-compose.yaml             # PostgreSQL、RabbitMQ、Redis、RustFS、Registry
├── templates/
│   ├── Dockerfile                      # Ubuntu 基础开发镜像
│   ├── Dockerfile.torch291              # PyTorch 2.9.1 镜像
│   ├── Dockerfile.torch291.codeserver   # PyTorch 2.9.1 + code-server
│   ├── Dockerfile.torch280.codeserver   # PyTorch 2.8.0 + code-server
│   ├── Dockerfile.mixm.codeserver       # mixm code-server 镜像
│   └── build.sh                        # 交互式构建与镜像打标
└── longhorn/longhorn.yaml               # Kubernetes 集群存储安装清单
```

## 启动顺序

应用部署文件统一位于 `deploy/web`，应用项目为 `mixm`，中间件项目为 `mixm-middleware`。五个中间件统一由 `deploy/middleware/docker-compose.yaml` 管理，两个项目通过外部 Docker 网络 `mixm-services` 互通。

先按[项目 README](../README.md)创建并编辑 `backend/.env`。Registry 默认保留原有的 `172.18.120.33:55000` 绑定；更换部署主机时设置 `REGISTRY_BIND_ADDRESS`、`REGISTRY_PORT`，并同步 `REGISTRY_URL` 和 `REGISTRY_PREFIX`。

使用 WSL + K3s + Longhorn 时，启动 mixm 前先按[集群存储说明](#kubernetes-集群存储)手动执行共享挂载设置脚本。

```bash
docker network inspect mixm-services >/dev/null 2>&1 || docker network create mixm-services

docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env up -d --wait
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env up -d --build backend frontend
```

PostgreSQL、RabbitMQ 和 Redis 使用健康检查；RustFS 和 Registry 的 `--wait` 等待容器进入运行状态。应用配置不声明跨项目 `depends_on`，因此需先启动中间件。本地 Go 开发也使用这一份中间件配置。

实验功能需要可用的 Kubernetes 配置；将 kubeconfig 放到 `backend/kubeconfig/k3s.yaml` 后，单独启动 worker：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env up -d --build worker
```

首次初始化数据与管理员的命令见[项目 README](../README.md)。后端启动时会自动执行 SQL 迁移；`003_remove_knowledge_tracing.sql` 删除已停用的掌握度与作答轨迹表，并清理历史测试中的掌握度和 Q 矩阵字段，试卷、答案和成绩保留。

## 中间件管理与连接

| 服务 | 项目 | 默认数据位置 | 宿主机端口 |
| --- | --- | --- | --- |
| PostgreSQL | `mixm-middleware` | 命名卷 `mixm_postgres_data` | `127.0.0.1:5432` |
| RabbitMQ | `mixm-middleware` | 命名卷 `mixm_rabbitmq_data` | `127.0.0.1:5672`、管理界面 `127.0.0.1:15672` |
| Redis | `mixm-middleware` | 命名卷 `mixm_redis_data` | `127.0.0.1:6379` |
| RustFS | `mixm-middleware` | 根目录 `RustFS/data` | S3 API `19000`、控制台 `19001` |
| Registry | `mixm-middleware` | 根目录 `registry/registry-data` | `172.18.120.33:55000`，可通过环境变量调整 |

查看状态、日志或单独管理一个服务：

```bash
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env ps
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env logs -f registry
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env stop redis
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env up -d --wait redis
```

停止两个项目：

```bash
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env down
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env down
```

应用的 `down` 只停止前端、后端和 worker；中间件的 `down` 停止五个中间件。普通 `down` 保留数据和共享网络；中间件的 `down -v` 会删除其管理的命名卷。

两个项目统一使用 `--env-file backend/.env`，保持数据库、消息队列和 RustFS 凭据一致。默认 PostgreSQL 数据库名、PostgreSQL 用户和 RabbitMQ 用户均为 `mixm`；分别用 `POSTGRES_DB`、`POSTGRES_USER`、`RABBITMQ_USER` 覆盖。本地 Go 开发修改这些值后，也需同步更新 `DATABASE_URL` 和 `RABBITMQ_URL`。

| 应用容器配置 | 默认地址 | 连接其他服务时的覆盖参数 |
| --- | --- | --- |
| `DATABASE_URL` | `postgres:5432`，数据库 `mixm` | `DEPLOY_DATABASE_URL` |
| `RABBITMQ_URL` | `rabbitmq:5672` | `DEPLOY_RABBITMQ_URL` |
| `REDIS_URL` | `redis:6379`，数据库 2 | `DEPLOY_REDIS_URL` |
| `RUSTFS_ENDPOINT` | `http://rustfs:9000` | `DEPLOY_RUSTFS_ENDPOINT` |
| `REGISTRY_URL` | `http://registry:5000` | `DEPLOY_REGISTRY_URL` |

`REGISTRY_URL` 在 `backend/.env` 中供本地 Go 使用；应用容器使用共享网络中的 `registry:5000`，通过 `DEPLOY_REGISTRY_URL` 覆盖。`REGISTRY_PREFIX` 是镜像名称中的仓库地址，必须能被构建主机与 Kubernetes 节点访问，默认与原 Registry 发布地址一致。改变发布端口时需同步调整本地 API 地址和镜像前缀。

共享桥接网络仅用于同机部署。PostgreSQL、RabbitMQ 和 Redis 默认只发布到宿主机 `127.0.0.1`；跨主机使用时，需调整端口绑定及应用的连接地址。

### 已有部署切换

合并只调整部署定义和模板位置，三个中间件命名卷及 RustFS、Registry 数据目录保持原位置。应用项目名仍为 `mixm`，中间件项目统一改为 `mixm-middleware`，新容器名由 Compose 自动生成。

切换前记录原容器、卷、数据库、账号和 RabbitMQ hostname：

```bash
docker ps -a
docker volume ls
docker inspect --format '{{.Config.Hostname}}' YOUR_EXISTING_RABBITMQ_CONTAINER
```

将 RabbitMQ hostname 写入 `backend/.env` 的 `RABBITMQ_HOSTNAME`，复用数据时保留原节点名，参阅[官方迁移说明](https://www.rabbitmq.com/docs/backup)。再停止原先分别部署的五个中间件容器，保留数据卷和目录，使用新配置启动。不要同时运行两个使用同一数据卷的实例。

如果现有卷名与默认值不同，使用以下参数指定实际旧卷名；这些参数不会复制或重命名数据：

| 参数 | 默认卷名 |
| --- | --- |
| `POSTGRES_VOLUME_NAME` | `mixm_postgres_data` |
| `RABBITMQ_VOLUME_NAME` | `mixm_rabbitmq_data` |
| `REDIS_VOLUME_NAME` | `mixm_redis_data` |
| `BACKEND_UPLOADS_VOLUME_NAME` | `mixm_backend_uploads` |
| `BACKEND_STATIC_VOLUME_NAME` | `mixm_backend_static` |
| `BACKEND_DATA_VOLUME_NAME` | `mixm_backend_data` |

初始化环境变量不会自动重命名已有 PostgreSQL 数据库、角色或 RabbitMQ 账号。复用旧数据库和账号时，保留实际名称与密码，或通过 `DEPLOY_*` 地址连接已有服务；实际账号改名需单独完成迁移。

Vite 默认允许的域名为 `mixm.yoobit.cn`，DNS 和反向代理按实际部署配置。配置改名不会修改外部域名资源。

## 应用镜像与实验模板

应用构建上下文分别为 `backend/` 和 `frontend/`：

```bash
docker build -f deploy/web/Dockerfile.backend -t mixm-backend backend
docker build -f deploy/web/Dockerfile.frontend -t mixm-frontend frontend
```

`deploy/web/Dockerfile.backend.dockerignore` 和 `deploy/web/Dockerfile.frontend.dockerignore` 分别作用于 `backend/` 与 `frontend/` 构建上下文，排除本地依赖、构建产物和环境文件。前端默认监听 `0.0.0.0:5173`；后端镜像提供 `server`、`worker`、`admin` 和 `initdb` 四个程序。

Ubuntu 模板镜像与构建脚本位于 `deploy/templates`。脚本根据自身位置选择构建上下文，可以从其他工作目录调用；通过交互输入或 `REGISTRY_PREFIX` 环境变量设置推送地址：

```bash
REGISTRY_PREFIX=172.18.120.33:55000 bash deploy/templates/build.sh
```

脚本构建并生成推送标签，推送命令在执行结束时显示。Dockerfile 统一使用 `docker/dockerfile:1`、大写指令、四空格续行缩进和 JSON 格式的 `CMD`；实验镜像使用 Bash `pipefail`，脚本 heredoc 保留运行时变量展开。

## Kubernetes 集群存储

Longhorn 作为独立的 Kubernetes 集群组件部署。

在 WSL + K3s 节点上，每次启动 mixm 前手动执行以下脚本，设置 Longhorn 所需的共享挂载传播，避免 `is not a shared mount` 错误：

```bash
sudo bash scripts/setup-longhorn-mount-propagation.sh
```

脚本仅执行 `mount --make-rshared /` 并显示根挂载的传播模式，预期为 `shared`。需要在 K3s 节点上以 root 权限执行，可以重复运行，无需重启 K3s。设置仅对当前运行的系统生效，重启系统或 WSL 后需重新执行；脚本不创建配置文件或开机服务。

已通过 Helm 安装 Longhorn 的集群继续使用 Helm 管理，无需再执行下面的安装清单：

```bash
kubectl apply -f deploy/longhorn/longhorn.yaml
bash scripts/create-demo-pod.sh demo-user
kubectl apply -f scripts/code-server-demo-user.yaml
```

演示脚本生成的 `scripts/code-server-<user_id>.yaml` 包含实例密码，已加入 Git 忽略规则。

## 配置校验

```bash
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env config --quiet
bash -n deploy/templates/build.sh
bash -n scripts/setup-longhorn-mount-propagation.sh
```

安装 Hadolint 后检查全部 7 个 Dockerfile；错误级检查失败时返回非零状态，版本固定等建议仍会显示：

```bash
hadolint --failure-threshold error \
  deploy/web/Dockerfile.backend \
  deploy/web/Dockerfile.frontend \
  deploy/templates/Dockerfile*
```

Docker daemon 可用时，还可以运行原生构建检查：

```bash
docker buildx build --check -f deploy/web/Dockerfile.backend backend
docker buildx build --check -f deploy/web/Dockerfile.frontend frontend
for dockerfile in deploy/templates/Dockerfile*; do
  docker buildx build --check -f "$dockerfile" deploy/templates
done
```
