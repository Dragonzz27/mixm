# REST API 与重构约定

业务接口使用 `/api`，此次直接改版，前后端必须同步发布。路由源位于 `internal/router`，完整清单为 `internal/router/routes.json`；运行服务后 `/openapi.json` 和 `/docs` 提供在线文档。

## 响应和分页

成功：`{"code":0,"message":"成功","data":...}`。失败：`{"code":HTTP状态码,"message":"错误说明","data":null}`，同时返回对应 HTTP 状态。查询、更新、删除返回 200，同步资源创建返回 201，实例任务受理返回 202。

分页参数为 `offset` 和 `limit`，默认 0 和 100；offset 不得为负，limit 为 1–1000。集合响应为 `data: {items,total,offset,limit}`。分类内容列表可携带 category 元数据；详情中的数组、分类树、权限集合和汇总结果保持业务结构。空集合返回空数组。PATCH 区分未提供、null、false、0 和空数组。

文件、SSE、`/health`、`/docs`、`/openapi.json` 不使用 JSON 包裹。OPTIONS 返回 204。请求支持末尾斜杠，不通过重定向匹配。普通 JSON 请求最多 8 MiB，且只能包含一个 JSON 值；Excel 导入、作业附件、内容上传上限分别为 20、50、100 MiB。

## 资源和工作流

常规资源使用 GET/POST 集合，以及 GET/PATCH/DELETE 单资源。路径 ID 是定位依据，提交和模板批量创建不再从请求体读取重复 ID。

| 功能 | 新接口（省略 /api） |
|---|---|
| 登录、刷新 | POST /auth/login、POST /auth/refresh |
| 用户、管理员创建用户 | /users、/users/:id；创建仍要求 Users 权限 |
| 密码、Excel 导入 | PUT /users/:id/password、POST /users/imports |
| 当前用户实例、镜像、提交 | GET /users/me/instances、/users/me/images、/users/me/submissions |
| 指定用户实例 | GET /users/:id/instances |
| 活动作业 | GET /assignments?active=true，仍检查截止时间 |
| 创建提交 | POST /assignments/:id/submissions |
| 查看提交、导出 | GET /assignments/:id/submissions、/assignments/:id/submissions/export |
| 提交评分、删除 | PATCH/DELETE /submissions/:id |
| 实例创建、启动、停止 | POST /instances；PATCH /instances/:id 的 status/replicas |
| 模板批量实例 | POST /templates/:id/instances，body 中保留 user_ids |
| 管理员删除全部实例 | DELETE /instances，返回逐项成功／失败报告 |
| 实例任务、汇总 | GET /instance-tasks、/instance-tasks/:id、/instance-tasks/summary |
| 绑定集合过滤 | GET /bindings?content_id=…&category_id=… |
| 分类批量绑定／解绑 | POST/DELETE /categories/:id/content-bindings，body: {content_ids:[…]} |
| 内容全部解绑、计数 | DELETE /contents/:id/category-bindings；GET /binding-counts?category_ids=… |
| 内容分类详情 | GET /contents/:id、GET /categories/:id?include=contents |
| 分类内容、未绑定内容 | GET /categories/:id/contents、GET /contents?unbound_category_id=… |
| 多分类内容 | GET /contents?category_ids=…，仍限定已发布且有效的内容 |
| 分类内容树、预览 | GET /categories/:id/content-tree、GET /categories?root_only=true&include=content-preview |
| 题目批量创建、审核 | POST /questions/batches、PUT /questions/:id/review |
| AI 题目、练习生成 | GET /questions?is_ai_generated=true、POST /practices/generations |
| 题目生成、知识树分析、保存 | POST /questions/generations、/knowledge-tree/analyses、/knowledge-tree |
| 模拟测试创建、交卷 | POST /simulation-attempts、POST /simulation-attempts/:id/submissions |
| 当前测试、解析、轨迹 | GET /simulation-attempts/current、/simulation-attempts/:id/review、/simulation-traces |
| 知识追踪状态 | GET /knowledge-tracing/status |
| 上传、下载 | POST /assignments/files、/contents/files；GET 对应 /files/:filename 或 /files/*key |
| 集群命名空间资源 | /clusters/namespaces/:namespace/pods/:name、deployments/:name、persistent-volume-claims/:name |
| PV/PVC 集合 | /clusters/persistent-volumes、/clusters/persistent-volume-claims |

保留 `/categories/tree`、`/question-categories/tree`、`/roles/permissions`、`/system/stats`、`/system/load` 等专用资源路径。历史 `/api/assignments/file/`、`/api/contents/file/`、`/static/` URL 继续可读，新上传返回新 URL。数据库和正文中的旧链接不做替换。

## 流式协议与前端

生成及分析接口返回 `text/event-stream`，事件仍为 `chunk/content`、`done/data`、`error/message`。流开始前的错误返回统一 JSON，开始后的错误通过 SSE 事件返回。前端共享 API_BASE_URL，普通请求解包 data，错误读取 message；刷新令牌使用独立请求以避免刷新递归。文件预览直接读取文件响应。

## 数据库和后台任务

GORM 映射现有表名、字符串主键、组合主键、nullable 字段、JSONB 和 TEXT JSON。保留物理删除、数据库默认值及显式更新时间。SQL 迁移位于 `internal/store/migrations`，保留名称、内容、应用记录和 advisory transaction lock；不使用 AutoMigrate。

任务状态与 outbox 保存在 PostgreSQL。发布保留确认，消费保留手动 ACK/NACK；worker 在一个固定连接上完成 advisory lock、处理与解锁，失败解锁时废弃连接。Redis 命中仍检查用户归属。Kubernetes 名称及用户 PVC 保留规则保持原行为。

测试仅使用独立 PostgreSQL schema、RabbitMQ 队列和 Redis 实例。测试环境变量为 TEST_DATABASE_URL、TEST_RABBITMQ_URL、TEST_REDIS_URL。执行 go test ./...、go vet ./...、make build 和前端 npm test / npm run build。
