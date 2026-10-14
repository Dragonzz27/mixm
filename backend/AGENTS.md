# Backend development

The active backend uses Go 1.25.0, net/http, pgx/PostgreSQL, Client-go, RabbitMQ and Redis. The Vue frontend consumes the existing /api routes. Python code under legacy/ is historical reference and is excluded from builds. kt_service is a separate model inference service.

Run from backend/:

- go test ./...
- go vet ./...
- make build
- go run ./cmd/server
- go run ./cmd/worker

Integration tests use TEST_DATABASE_URL, TEST_RABBITMQ_URL and TEST_REDIS_URL. Use disposable test services. Tests create independent PostgreSQL schemas and RabbitMQ queues and clean them up.

Use gofmt, context-aware I/O, parameterized SQL, pgx transactions for related writes, and meaningful HTTP errors with the existing detail JSON key. Runtime persistence must use PostgreSQL. SQLite belongs only in cmd/migrate-sqlite, which opens the source read-only.

Keep API routes and frontend payload fields compatible. legacy_routes.json and TestLegacyRouteCoverage cover the original route inventory. Authorization must be enforced in handlers, and task ownership must be checked before using Redis cache entries.

Keep durable task records and the publishing outbox in PostgreSQL. RabbitMQ messages contain task IDs, use publisher confirmations and manual consumer acknowledgement, and must remain safe under redelivery. Preserve deploy-cs-, svc-cs-, ing-cs-, mid-cs- and edu-pvc-cs- resource names for migrated instances. User PVCs survive instance deletion.

Migrations are embedded in internal/store/migrations, applied in filename order, and serialized by an advisory transaction lock. Do not rewrite applied migrations; add a new numbered migration.

Do not log credentials or include .env files, kubeconfigs, SSH host keys or database files in images. Never run imports or schema changes against existing user databases as part of tests.

Deployment files live under deploy/ at the repository root. The main entry point is docker compose -f deploy/docker-compose.yaml; Dockerfiles are selected explicitly while build contexts remain backend/, frontend/ and kt_service/.
