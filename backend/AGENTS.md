# Backend development

The active backend uses Go 1.25.0, Gin, GORM/PostgreSQL, Client-go, RabbitMQ and Redis. The Vue frontend consumes the existing /api routes.

Run from backend/:

- go test ./...
- go vet ./...
- make build
- go run ./cmd/server
- go run ./cmd/worker

Integration tests use TEST_DATABASE_URL, TEST_RABBITMQ_URL and TEST_REDIS_URL. Use disposable test services. Tests create independent PostgreSQL schemas and RabbitMQ queues and clean them up.

Use gofmt, context-aware I/O, parameterized SQL, GORM transactions for related writes, and meaningful HTTP errors. Ordinary JSON responses use code/message/data; code is zero on success and the HTTP status on failure. Persistence uses PostgreSQL. This is a new project; retain PostgreSQL schema initialization and application seed commands.

Keep code readable: put each logical statement on its own line, expand long composite literals and test case tables, and separate functions and business phases with blank lines. Group imports as standard library, external dependencies, then mixm/backend packages. Add brief Chinese comments above functions and meaningful business blocks, normally one or two sentences explaining their purpose or constraints. Keep error handling with its related operation and avoid comments that merely repeat individual assignments. Run gofmt after editing.

The REST contract lives in internal/router/routes.json and docs/api.md. Keep backend and frontend changes coordinated. Lists use items/total/offset/limit, partial updates use PATCH, and path IDs locate resources. JWT authentication belongs in middleware; services enforce permissions and ownership. Task ownership must be checked before using Redis cache entries.

Organize HTTP routing, middleware, handlers, DTOs, business services, repositories, models, external clients and bootstrap code separately. Handlers do not contain SQL or business rules; services do not receive gin.Context, HTTP requests or response writers. Repository queries use GORM and explicitly map the existing schema. Do not use AutoMigrate or soft deletion. PATCH DTOs must preserve absent fields, explicit null and zero values.

Infrastructure packages live under internal/infra/{client,cluster,queue,ssh}. The infra directory is only a grouping directory, not a shared Go package. Keep package-local helpers with their callers; do not introduce a catch-all utils package. Shared code used by the four commands belongs in internal; do not introduce backend/pkg without a concrete cross-project consumer. Application errors belong in internal/apperror, not dto; external clients must not depend on DTOs for error reporting. The auth package owns JWT and password hashing, while middleware owns HTTP authentication.

The repository package owns database connection setup, numbered migrations, queries, transactions and pinned task locks. Bootstrap code uses repository.Open and repository.Migrate. Keep connection setup, migrations, query binding and task locks in separate files within that package. Queue transport, publishing, consumption and task execution likewise use separate files within the queue package.

Worker task locks are session-scoped: pin one physical connection through lock, processing and unlock, and discard it if unlocking fails. Preserve the SQL migration advisory transaction lock and the RabbitMQ outbox transaction boundary.

Keep historical /api/assignments/file/, /api/contents/file/ and /static/ download URLs working. JSON envelopes do not apply to files, SSE, /health, /docs or /openapi.json. Before starting SSE, failures use the JSON envelope; after streaming starts, send error events.

Keep durable task records and the publishing outbox in PostgreSQL. RabbitMQ messages contain task IDs, use publisher confirmations and manual consumer acknowledgement, and must remain safe under redelivery. Preserve deploy-cs-, svc-cs-, ing-cs-, mid-cs- and edu-pvc-cs- resource names for instances. User PVCs survive instance deletion.

Migrations are embedded in internal/repository/migrations, applied in filename order, and serialized by an advisory transaction lock. Do not rewrite applied migrations; add a new numbered migration.

Do not log credentials or include .env files, kubeconfigs, SSH host keys or database files in images. Never run imports or schema changes against existing user databases as part of tests.

Deployment files live under deploy/ at the repository root. The application entry point is docker compose -f deploy/web/docker-compose.yaml and contains only frontend, backend and worker. PostgreSQL, RabbitMQ, Redis, RustFS and Registry are defined together in deploy/middleware/docker-compose.yaml and share the external mixm-services network with backend and worker. Experiment image Dockerfiles and their build script live in deploy/templates/. Application Dockerfiles are deploy/web/Dockerfile.backend and deploy/web/Dockerfile.frontend, with matching Dockerfile-specific .dockerignore files; build contexts remain backend/ and frontend/.
