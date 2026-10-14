# 部署文件

所有部署命令默认从项目根目录执行。

```text
deploy/
├── docker-compose.yaml                 # 完整应用与 PostgreSQL、RabbitMQ、Redis、RustFS
├── backend/
│   ├── Dockerfile                      # Go server、worker 和管理命令
│   └── Dockerfile.dockerignore         # 后端构建上下文排除规则
├── frontend/Dockerfile                 # Vue 开发服务镜像
├── kt-service/Dockerfile               # 独立模型推理服务镜像
├── registry/
│   ├── docker-compose.yaml            # 独立镜像仓库
│   └── ubuntu/                        # 实验镜像 Dockerfile 和 build.sh
├── rustfs/docker-compose.yaml          # 可选的独立 RustFS 部署
└── longhorn/longhorn.yaml               # Longhorn 安装清单
```

## 完整应用

```bash
cp backend/.env.example backend/.env
# 配置 backend/.env，并将 kubeconfig 放到 backend/kubeconfig/k3s.yaml
docker compose -f deploy/docker-compose.yaml --env-file backend/.env up -d --build
docker compose -f deploy/docker-compose.yaml --env-file backend/.env logs -f backend worker
```

首次初始化和旧数据库迁移参阅 [项目 README](../README.md) 与 [后端迁移说明](../backend/docs/migration.md)。仅启动本地 Go 开发依赖：

```bash
docker compose -f deploy/docker-compose.yaml up -d postgres rabbitmq redis
```

Dockerfile 位于此目录，构建上下文仍为原应用目录，因此 `COPY` 路径保持原样。单独构建镜像：

```bash
docker build -f deploy/backend/Dockerfile -t mixm-backend backend
docker build -f deploy/frontend/Dockerfile -t mixm-frontend frontend
docker build -f deploy/kt-service/Dockerfile -t mixm-kt-service kt_service
```

后端排除规则采用与 Dockerfile 同目录的 `Dockerfile.dockerignore`，匹配的路径相对于 `backend/` 构建上下文。

## 独立基础服务

```bash
docker compose -f deploy/registry/docker-compose.yaml up -d
docker compose -f deploy/rustfs/docker-compose.yaml up -d
```

完整应用已包含 RustFS；独立 RustFS 配置用于单独部署。Registry 保留原有的主机 IP 和端口绑定，可按部署主机修改 `deploy/registry/docker-compose.yaml`。

移动配置后，数据位置与 Compose 项目名称保持一致：

| 配置 | Compose 项目名 | 宿主机数据目录 |
| --- | --- | --- |
| 完整应用 | `mixm-edu` | 原有命名卷，以及根目录 `RustFS/data` |
| 独立 Registry | `registry` | 根目录 `registry/registry-data` |
| 独立 RustFS | `rustfs` | 根目录 `RustFS/data` |

挂载的 kubeconfig、前端源码和模型输出仍使用原有目录。应用 `.env.example` 保留在应用目录，供本地开发与部署共享；运行数据不随部署文件迁移。

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
docker compose -f deploy/docker-compose.yaml config --quiet
docker compose -f deploy/registry/docker-compose.yaml config --quiet
docker compose -f deploy/rustfs/docker-compose.yaml config --quiet
```
