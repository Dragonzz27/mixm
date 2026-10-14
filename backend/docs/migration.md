# Go 后端迁移说明

## 当前实现

运行环境固定为 Go 1.25.0。HTTP 和 SSH 网关由 server 提供，worker 通过 RabbitMQ 创建实验实例。使用 pgx 连接 PostgreSQL，Redis 保存任务终态缓存。Python/Celery/SQLite 已退出主后端运行链路；SQLite 驱动只被数据导入命令引用，kt_service 继续作为独立的模型服务。

数据库迁移在 server、worker 和管理命令启动时执行。SQL 文件嵌入二进制，迁移事务通过 PostgreSQL advisory lock 串行执行，schema_migrations 防止重复执行。关联更新、模拟测试评分和任务创建使用事务。

## 任务投递与恢复

1. HTTP 将实例、随机密码和任务 payload 一起持久化到 PostgreSQL，返回 task_id 和 instance_id。
2. outbox publisher 使用 FOR UPDATE SKIP LOCKED 获取任务，通过 RabbitMQ durable queue、persistent message 和 publisher confirm 投递。
3. 收到确认后记录 published_at；未确认任务保留并重试。pending/processing 任务超过十分钟可再次投递，以恢复遗失投递。
4. worker 使用 PostgreSQL 会话 advisory lock 锁定任务，防止多个消费者同时创建相同资源；固定的实例 ID 和 Kubernetes 资源名保证重复投递幂等。
5. 等待 Deployment 就绪，再提交实例和任务状态，随后 ACK。Redis 缓存写入失败不影响 PostgreSQL 的终态。终态缓存只向任务所属用户开放。
6. 已提交终态的重复消息直接 ACK。进程中断时，未确认消息由 RabbitMQ 重新投递；失败任务保存错误，可由用户重新创建。

Redis 是可丢弃缓存，RabbitMQ 是任务传输通道，PostgreSQL 是任务状态的持久来源。

## 接口与行为

134 条旧接口路径和方法由 legacy_routes.json 与自动测试覆盖。保留登录返回的 user/routes/menu、JWT access/refresh、bcrypt 密码、角色可见性、附件 URL、SSH 连接字段及模拟测试返回结构。

有意调整的行为：

- POST /api/instances/ 与 /api/instances/start 均返回 202 和任务信息；创建由 worker 完成。
- 用户管理和教学资源写入需要登录及对应菜单权限；批量实例操作仍要求超级管理员。刷新令牌不能用作访问令牌；用户名 admin 不会自动获得超级管理员权限。
- Kubernetes 操作失败不再把数据库实例状态写成成功。任务创建中的实例暂停启动/删除操作，避免与 worker 并发改变资源。
- 模拟测试交卷读取数据库保存的试卷及答案，不信任请求中的 questions 分数和正确答案；重复提交返回 409。
- 新增分页上限、资源参数校验、唯一约束及外键。旧数据不满足约束时，导入会回滚并报告表或字段。

API 文档的 OpenAPI 自动列出接口、路径参数和 Bearer 认证。复杂请求字段以原有前端请求、Go 校验逻辑和集成测试为准。

## 切换已有安装

1. 停止旧系统新增写入，等待 Celery 任务完成。备份 SQLite 文件、uploads、RustFS 对象及 kubeconfig，记录运行中的 Kubernetes 实例。
2. 配置新的 PostgreSQL/RabbitMQ/Redis 和同一个集群、Registry、RustFS。不要删除原有 PVC 或 S3 bucket。
3. 对 SQLite 导入运行 dry-run，处理重复绑定、多角色用户、多次作业提交等与新唯一约束冲突的数据，再正式导入到空业务数据库。
4. 沿用原 SECRET_KEY 可验证尚未过期的旧访问令牌；保留 email/username/bcrypt 密码，原账号可重新登录。Celery 消息和 Redis 任务历史不自动转换。
5. 将原 backend/uploads 内容复制到 Compose 的 backend_uploads 卷，并保持 uploads/assignments 相对路径。原成绩导出文件需要时复制到 backend_static 卷；镜像内使用 uid/gid 10001，保证该用户可读写对应文件。
6. 将 kubeconfig 放到 backend/kubeconfig/k3s.yaml，确保 server 地址不是容器中的 127.0.0.1，引用的证书文件也可访问；或使用集群内 ServiceAccount。
7. 启动新 server/worker，确认用户登录、角色菜单、内容、附件、作业、模拟测试，以及实例创建/停止/启动/删除。

资源名称沿用 deploy-cs-{instance_id}、svc-cs-{instance_id}、ing-cs-{instance_id}、mid-cs-{instance_id}、edu-pvc-cs-{user_id}。删除实例保留用户 PVC，重新创建实例继续挂载该用户原有 workspace。Longhorn 卷延续原有行为，副本数设为 1；支持 v1beta2/v1beta1 API。

code-server 的 /instance/{id} 路由需要 Traefik ingress controller 和 traefik.io/v1alpha1 Middleware CRD。SSH 网关使用 Client-go pods/exec，ServiceAccount 必须有对应权限；主机密钥写入持久的 backend_data 卷。

## 故障检查

```bash
docker compose logs backend worker postgres rabbitmq
```

RabbitMQ 无法连接时 outbox 保留任务并自动重连。缺少 kubeconfig 时教学接口仍可运行，集群/实例接口返回 503，worker 启动失败。KT 推理不可用时沿用旧版启发式掌握度更新（答对 alpha=0.18、答错 beta=0.22）。LLM 接口需要有效 LLM_API_KEY 和兼容的 /chat/completions 服务。

## 本次验证范围

已验证 Go 编译、单元测试、134 条路由覆盖、PostgreSQL API 流程、RabbitMQ 消费/重复消息、Redis 缓存及合成 SQLite 数据的 dry-run/正式导入。Kubernetes 资源测试使用 fake client。真实集群、Traefik、SSH 终端、RustFS、Registry、LLM 和模型 checkpoint 仍需在部署环境进行联调。
