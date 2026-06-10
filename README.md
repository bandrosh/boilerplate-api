# boilerplate-api

REST API boilerplate in **Go**, built with a **Hexagonal (Ports & Adapters) + DDD-lite** architecture, ready for robust production projects.

The application **runs from the IDE / locally** (`go run ./cmd/api`); **Docker contains infrastructure only** (LocalStack/DynamoDB + an observability stack with Grafana).

## Stack

| Layer            | Technology                                              |
|------------------|--------------------------------------------------------|
| HTTP router      | [chi](https://github.com/go-chi/chi)                   |
| Persistence      | **DynamoDB** ([AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)) — single-table design |
| Local AWS        | [LocalStack](https://localstack.cloud)                 |
| Config           | env + `.env` ([caarlos0/env](https://github.com/caarlos0/env), godotenv) |
| Logging          | `log/slog` (structured)                                |
| Observability    | OpenTelemetry → OTEL Collector → Prometheus / Tempo / Grafana |

## Architecture

```
cmd/api                     # composition root (main)
internal/
  domain/user              # entities, value objects, errors, PORTS (interfaces)
  application/user         # use cases (orchestration)
  adapter/
    http                   # inbound adapter (chi): router, handlers, dto, middleware
    dynamodb               # outbound adapter: AWS SDK v2 client, single-table repository
  platform/                # cross-cutting: config, logger, observability, server
docker/                    # infra configs (localstack init, otel, prometheus, tempo, grafana)
```

The dependency rule points **inward**: `adapter → application → domain`. The domain knows nothing about HTTP or DynamoDB; adapters implement the *ports* defined in the domain. Swapping persistence (e.g. DynamoDB → Postgres) only affects the outbound adapter.

### DynamoDB modeling (single-table design)

A single table (`boilerplate`) with generic `PK`/`SK` keys and a global index `GSI1` (`GSI1PK`/`GSI1SK`):

| Item        | PK              | SK              | GSI1PK | GSI1SK              |
|-------------|-----------------|-----------------|--------|--------------------|
| User        | `USER#<id>`     | `USER#<id>`     | `USER` | `<createdAt>#<id>` |
| Email lock  | `EMAIL#<email>` | `EMAIL#<email>` | –      | –                  |

- **Email uniqueness:** creation writes the user item and the email-lock item in a single `TransactWriteItems` with `attribute_not_exists`, guaranteeing atomic uniqueness.
- **Listing:** `Query` on `GSI1` (`GSI1PK = USER`), ordered by date, with **cursor-based pagination** (`next_cursor`).

## Running

### 1. Prerequisites
- Go 1.23+
- Docker + Docker Compose

### 2. Start the infrastructure
```bash
make infra-up        # localstack(dynamodb) + otel-collector + prometheus + tempo + grafana
```
LocalStack creates the table automatically via the init script (`docker/localstack/init`).
Check it with `make db-tables`.

### 3. Configure the environment
```bash
cp .env.example .env
```

### 4. Run the API (from the IDE or terminal)
```bash
make run             # go run ./cmd/api
```

The API listens on `http://127.0.0.1:8080`.

## Endpoints

| Method | Route                 | Description                        |
|--------|-----------------------|------------------------------------|
| GET    | `/healthz`            | Liveness probe                     |
| GET    | `/readyz`             | Readiness (checks DynamoDB)        |
| POST   | `/api/v1/users`       | Create user                        |
| GET    | `/api/v1/users`       | List (`?limit=&cursor=`)           |
| GET    | `/api/v1/users/{id}`  | Get by id                          |
| PUT    | `/api/v1/users/{id}`  | Update name                        |
| DELETE | `/api/v1/users/{id}`  | Delete                             |

Example:
```bash
curl -X POST 127.0.0.1:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","email":"ada@example.com"}'

# Paginated listing
curl "127.0.0.1:8080/api/v1/users?limit=10"
# => {"items":[...],"next_cursor":"..."}  (pass the cursor for the next page)
```

## Services (UI / observability)

> On Windows/macOS use **`127.0.0.1`**, not `localhost`: `localhost` resolves to IPv6 (`::1`) while Docker Desktop's published ports only answer on IPv4, which causes browser timeouts.

| Service         | URL                       | Login         |
|-----------------|---------------------------|---------------|
| Grafana         | http://127.0.0.1:3000     | admin / admin |
| Prometheus      | http://127.0.0.1:9090     | –             |
| Tempo (API)     | http://127.0.0.1:3200     | –             |
| Jaeger UI       | http://127.0.0.1:16686    | –             |
| DynamoDB Admin  | http://127.0.0.1:8001     | –             |
| LocalStack      | http://127.0.0.1:4566     | –             |

The application exports traces and metrics via OTLP (`127.0.0.1:4317`) to the OTEL Collector, which fans out to Tempo and Jaeger (traces) and Prometheus (metrics). Grafana ships with provisioned datasources and an HTTP dashboard. **DynamoDB Admin** lets you browse the DynamoDB tables in LocalStack.

## Useful commands

```bash
make help            # list all targets
make db-tables       # list DynamoDB tables (LocalStack)
make db-scan         # scan the table (debug)
make test            # unit tests (-race and coverage)
make test-integration # integration tests (requires infra: make infra-up)
make lint            # golangci-lint
make infra-down      # tear down the infra
```

## Tests

- **Unit:** domain (`internal/domain/user`) and use cases (`internal/application/user`, with a fake repository). Always run: `make test`.
- **Integration:** DynamoDB repository against LocalStack, behind the `integration` build tag. Start the infra (`make infra-up`) and run `make test-integration`. CI has a dedicated job with LocalStack.

## Adding a new feature

1. Model the domain in `internal/domain/<feature>` (entity + port).
2. Write the use cases in `internal/application/<feature>`.
3. Implement the repository in `internal/adapter/dynamodb` (define the key pattern in the single table).
4. Expose it via handlers in `internal/adapter/http` and mount it in `router.go`.
5. If you need new tables/indexes, adjust the init in `docker/localstack/init`.
