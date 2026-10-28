# 部署文件

所有部署命令默认从项目根目录执行。

Compose 和 Dockerfile 中的基础镜像默认从 Docker Hub 拉取。RustFS 使用项目维护的官方镜像 `rustfs/rustfs:1.0.4`，固定版本避免随 `latest` 标签更新；PostgreSQL、RabbitMQ、Redis、Registry、Node.js、Go、Alpine 和 Ubuntu 使用 Docker Official Images。

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
└── templates/
    ├── Dockerfile.mixm                       # 基础开发环境 + SSH
    ├── Dockerfile.mixm.codeserver            # 基础开发环境 + SSH + code-server
    ├── Dockerfile.mixm.codeserver.torch280   # code-server + PyTorch 2.8.0
    ├── Dockerfile.mixm.codeserver.torch291   # code-server + PyTorch 2.9.1
    ├── Dockerfile.mixm.codeserver.wsl2       # WSL2 / HAMi GPU 测试镜像
    ├── scripts/                             # 共用用户配置、虚拟环境和启动脚本
    └── build.sh                             # 交互式构建与镜像打标
```

## 启动顺序

应用部署文件统一位于 `deploy/web`，应用项目为 `mixm`，中间件项目为 `mixm-middleware`。五个中间件统一由 `deploy/middleware/docker-compose.yaml` 管理，两个项目通过外部 Docker 网络 `mixm-services` 互通。

先按[项目 README](../README.md)创建并编辑 `backend/.env`。环境示例默认用于本地开发，Registry 绑定 `127.0.0.1:55000`，通过 `localhost:55000` 访问；跨主机部署时设置 `REGISTRY_BIND_ADDRESS`、`REGISTRY_PORT`，并将 `REGISTRY_URL` 和 `REGISTRY_PREFIX` 调整为相关主机与 Kubernetes 节点可访问的地址。

使用 WSL + K3s + Longhorn 时，启动 mixm 前先按[集群存储说明](#kubernetes-集群存储)手动执行共享挂载设置脚本。

首次启动前，以当前开发用户创建挂载目录，再启动服务：

```bash
mkdir -p .mixm/{rustfs,registry,backend}

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

## 持久化目录

开发数据统一放在项目根目录的 `.mixm/` 下，已加入 Git 忽略规则：

```text
.mixm/
├── rustfs/                       # RustFS 对象数据，默认桶 content
│   └── logs/                     # RustFS 日志
├── registry/                     # Registry 镜像数据
└── backend/
    ├── uploads/assignments/       # 作业附件
    ├── static/export/             # 导出文件
    └── data/ssh_host_key          # SSH Gateway 主机私钥
```

首次启动前，从项目根目录创建三个挂载目录即可，无需额外脚本或 sudo，可重复执行：

```bash
mkdir -p .mixm/{rustfs,registry,backend}
```

后端按需创建上传、导出和 SSH 私钥目录；SSH 私钥目录权限为 `0700`，私钥文件权限为 `0600`。RustFS 自行创建日志子目录。提前创建挂载目录可避免 Docker 自动创建为 root 所有，导致按开发用户 UID 运行的容器无法写入。

后端文件路径默认以 `../.mixm/backend` 开头，因此本地 Go 命令需在 `backend/` 下执行。后端容器的工作目录为 `/app/backend`，对应的数据挂载到 `/app/.mixm/backend`。

RustFS、Registry、backend 和 worker 容器在 Compose 中默认使用 `1000:1000`，无需在 `.env` 中重复配置。若宿主机开发用户的 UID/GID 不同，可通过 `MIXM_UID`、`MIXM_GID` 覆盖，保证容器与本地 Go 进程共享目录时都能写入。RustFS 日志写入 `/data/logs`，对应 `.mixm/rustfs/logs`。

PostgreSQL、RabbitMQ 和 Redis 使用 Docker 命名卷；实验实例的 `/workspace` 由 Kubernetes PVC 和 Longhorn 管理。

## 中间件管理与连接

| 服务 | 项目 | 默认数据位置 | 宿主机端口 |
| --- | --- | --- | --- |
| PostgreSQL | `mixm-middleware` | 命名卷 `mixm_postgres_data` | `127.0.0.1:5432` |
| RabbitMQ | `mixm-middleware` | 命名卷 `mixm_rabbitmq_data` | `127.0.0.1:5672`、管理界面 `127.0.0.1:15672` |
| Redis | `mixm-middleware` | 命名卷 `mixm_redis_data` | `127.0.0.1:6379` |
| RustFS | `mixm-middleware` | `.mixm/rustfs` | S3 API `19000`、控制台 `19001` |
| Registry | `mixm-middleware` | `.mixm/registry` | `127.0.0.1:55000`，可通过环境变量调整 |

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

应用的 `down` 只停止前端、后端和 worker；中间件的 `down` 停止五个中间件。普通 `down` 保留数据和共享网络；中间件的 `down -v` 会删除其管理的命名卷。`.mixm` 下的绑定目录不会被 `down -v` 删除。

两个项目统一使用 `--env-file backend/.env`，保持数据库、消息队列和 RustFS 凭据一致。默认 PostgreSQL 数据库名、PostgreSQL 用户和 RabbitMQ 用户均为 `mixm`；分别用 `POSTGRES_DB`、`POSTGRES_USER`、`RABBITMQ_USER` 覆盖。本地 Go 开发修改这些值后，也需同步更新 `DATABASE_URL` 和 `RABBITMQ_URL`。

| 应用容器配置 | 默认地址 | 连接其他服务时的覆盖参数 |
| --- | --- | --- |
| `DATABASE_URL` | `postgres:5432`，数据库 `mixm` | `DEPLOY_DATABASE_URL` |
| `RABBITMQ_URL` | `rabbitmq:5672` | `DEPLOY_RABBITMQ_URL` |
| `REDIS_URL` | `redis:6379`，数据库 2 | `DEPLOY_REDIS_URL` |
| `RUSTFS_ENDPOINT` | `http://rustfs:9000` | `DEPLOY_RUSTFS_ENDPOINT` |
| `REGISTRY_URL` | `http://registry:5000` | `DEPLOY_REGISTRY_URL` |

`REGISTRY_URL` 在 `backend/.env` 中供本地 Go 使用，默认值为 `http://localhost:55000`；应用容器使用共享网络中的 `registry:5000`，通过 `DEPLOY_REGISTRY_URL` 覆盖。`REGISTRY_PREFIX` 是镜像名称中的仓库地址，默认值为 `localhost:55000`，用于同机开发。跨主机或多节点集群需改为构建主机与各 Kubernetes 节点均可访问的地址，并调整 Registry 端口绑定。改变发布端口时需同步调整本地 API 地址和镜像前缀。

共享桥接网络仅用于同机部署。PostgreSQL、RabbitMQ 和 Redis 默认只发布到宿主机 `127.0.0.1`；跨主机使用时，需调整端口绑定及应用的连接地址。

## 应用镜像与实验模板

应用构建上下文分别为 `backend/` 和 `frontend/`：

```bash
docker build -f deploy/web/Dockerfile.backend -t mixm-backend backend
docker build -f deploy/web/Dockerfile.frontend -t mixm-frontend frontend
```

`deploy/web/Dockerfile.backend.dockerignore` 和 `deploy/web/Dockerfile.frontend.dockerignore` 分别作用于 `backend/` 与 `frontend/` 构建上下文，排除本地依赖、构建产物和环境文件。前端默认监听 `0.0.0.0:5173`；后端镜像提供 `server`、`worker`、`admin` 和 `initdb` 四个程序。

实验模板均基于 Docker Hub 的 `ubuntu:24.04`，可分别独立构建，无需先构建其他模板镜像。共用脚本集中在 `deploy/templates/scripts/`；构建上下文必须为 `deploy/templates`，以便复制这些脚本。

| 模板 | Python 环境 | code-server | PyTorch / torchvision | 默认计算设备 |
| --- | --- | --- | --- | --- |
| `Dockerfile.mixm` | uv、`/opt/venv` | 无 | 无 | — |
| `Dockerfile.mixm.codeserver` | uv、`/opt/venv` | 有 | 无 | — |
| `Dockerfile.mixm.codeserver.torch280` | uv、`/opt/venv` | 有 | `2.8.0` / `0.23.0` | CPU |
| `Dockerfile.mixm.codeserver.torch291` | uv、`/opt/venv` | 有 | `2.9.1` / `0.24.1` | CPU |
| `Dockerfile.mixm.codeserver.wsl2` | uv、`/opt/venv` | 有 | 无 | — |

PyTorch 与 torchvision 版本按[官方兼容组合](https://pytorch.org/get-started/previous-versions/)固定。所有模板统一基础开发工具、用户 `ubuntu`、工作目录 `/workspace` 和 SSH 配置；code-server 使用全局 standalone 安装，默认监听 `8080`，禁用更新检查。SSH 登录和交互式 Bash 自动使用虚拟环境。普通 Python 包使用 uv 默认的官方 PyPI，PyTorch 使用官方 `download.pytorch.org`。

构建脚本根据自身位置选择构建上下文，可以从其他工作目录调用。菜单对应上述五个模板，两个 `torch280`、`torch291` 模板会询问 CPU/CUDA，选择 CUDA 时再设置通道（默认 `cu128`）；选项 5 为基础 code-server 的 WSL2 测试镜像，不安装 PyTorch 或 CUDA。通过交互输入或 `REGISTRY_PREFIX` 环境变量设置推送地址：

```bash
REGISTRY_PREFIX=localhost:55000 bash deploy/templates/build.sh
```

默认镜像名称由模板名转换，例如 `mixm-codeserver-torch291`。基础模板与通用 code-server 模板默认标签为 `v0.0.1`，PyTorch 模板按计算设备使用 `cpu-v0.0.1` 或 `cu128-v0.0.1` 等标签；WSL2 测试镜像默认为 `mixm-codeserver-wsl2:v0.0.1`。脚本构建并生成 Registry 标签，推送命令在执行结束时显示，推送需自行执行。

也可以直接构建和推送，例如 PyTorch 2.9.1 + CUDA 12.8 通道：

```bash
docker build \
  -f deploy/templates/Dockerfile.mixm.codeserver.torch291 \
  --build-arg TORCH_DEVICE=cuda \
  --build-arg TORCH_CUDA_TAG=cu128 \
  -t localhost:55000/mixm-codeserver-torch291:cu128-v0.0.1 \
  deploy/templates

docker push localhost:55000/mixm-codeserver-torch291:cu128-v0.0.1
```

WSL2 + HAMi 测试使用独立的 `Dockerfile.mixm.codeserver.wsl2`，内容沿用 `Dockerfile.mixm.codeserver`，仅补充预加载变量及用户切换后的恢复逻辑，不安装 PyTorch 或 CUDA。该镜像内置以下变量，并在用户切换后通过虚拟环境加载脚本恢复，无需后端配置或 Pod 环境变量注入：

```dockerfile
ENV LD_PRELOAD=/usr/local/vgpu/libvgpu.so:/usr/lib/wsl/drivers/nv_dispi.inf_amd64_b20cc8aeaed64fc2/libcuda.so.1.1
```

构建并推送：

```bash
docker build \
  -f deploy/templates/Dockerfile.mixm.codeserver.wsl2 \
  -t localhost:55000/mixm-codeserver-wsl2:v0.0.1 \
  deploy/templates

docker push localhost:55000/mixm-codeserver-wsl2:v0.0.1
```

在前端创建仓库记录 `mixm-codeserver-wsl2`，创建实例时选择该镜像标签，并申请 GPU 和单卡显存。共享库由 HAMi 和 NVIDIA 容器运行环境在启动时提供，不复制进镜像；这个测试镜像固定了当前 WSL2 驱动路径，驱动目录变化后需修改 Dockerfile 并重新构建。进入 code-server 终端后可执行 `printf '%s\n' "$LD_PRELOAD"` 检查预加载变量。

Dockerfile 统一使用 `docker/dockerfile:1`、大写指令、四空格续行缩进和 JSON 格式的 `CMD`；实验镜像使用 Bash `pipefail`。

## Kubernetes 集群存储

Longhorn 作为独立的 Kubernetes 集群组件，通过 Helm 安装和管理。

Windows 浏览器访问 WSL 内的实例时，若 `localhost:80` 无法连接，但 WSL 节点 IP 的 `80` 端口可以打开 code-server 登录页，将本地 `backend/.env` 的 `K8S_CODE_SERVER_HOST` 设置为该节点 IP。可通过 `kubectl get nodes -o wide` 查看 `INTERNAL-IP`。修改后重启 Go API 并刷新实例详情页，已有实例的访问地址也会更新。WSL 节点 IP 改变后需同步修改；`.env.example` 保留 `localhost` 开发默认值。

在 WSL + K3s 节点上，每次启动 mixm 前手动执行以下脚本，设置 Longhorn 所需的共享挂载传播，避免 `is not a shared mount` 错误：

```bash
sudo bash scripts/setup-longhorn-mount-propagation.sh
```

脚本仅执行 `mount --make-rshared /` 并显示根挂载的传播模式，预期为 `shared`。需要在 K3s 节点上以 root 权限执行，可以重复运行，无需重启 K3s。设置仅对当前运行的系统生效，重启系统或 WSL 后需重新执行；脚本不创建配置文件或开机服务。

Longhorn 就绪后，可生成并部署演示实例：

```bash
bash scripts/create-demo-pod.sh demo-user
kubectl apply -f scripts/code-server-demo-user.yaml
```

演示脚本生成的 `scripts/code-server-<user_id>.yaml` 包含实例密码，已加入 Git 忽略规则。

## 配置校验

```bash
docker compose -f deploy/middleware/docker-compose.yaml --env-file backend/.env config --quiet
docker compose -f deploy/web/docker-compose.yaml --env-file backend/.env config --quiet
bash -n deploy/templates/build.sh
for script in deploy/templates/scripts/*.sh; do
  bash -n "$script"
done
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
