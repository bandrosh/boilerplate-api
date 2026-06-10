# boilerplate-api

Boilerplate de API REST em **Go**, com arquitetura **Hexagonal (Ports & Adapters) + DDD-lite**, pronto para projetos robustos de produção.

A aplicação **roda pela IDE/localmente** (`go run ./cmd/api`); o **Docker contém apenas a infraestrutura** (LocalStack/DynamoDB + stack de observabilidade com Grafana).

## Stack

| Camada           | Tecnologia                                              |
|------------------|--------------------------------------------------------|
| HTTP router      | [chi](https://github.com/go-chi/chi)                   |
| Persistência     | **DynamoDB** ([AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)) — single-table design |
| AWS local        | [LocalStack](https://localstack.cloud)                 |
| Config           | env + `.env` ([caarlos0/env](https://github.com/caarlos0/env), godotenv) |
| Logging          | `log/slog` (estruturado)                               |
| Observabilidade  | OpenTelemetry → OTEL Collector → Prometheus / Tempo / Grafana |

## Arquitetura

```
cmd/api                     # composition root (main)
internal/
  domain/user              # entidades, value objects, erros, PORTS (interfaces)
  application/user         # use cases (orquestração)
  adapter/
    http                   # inbound adapter (chi): router, handlers, dto, middleware
    dynamodb               # outbound adapter: client AWS SDK v2, repositório single-table
  platform/                # cross-cutting: config, logger, observability, server
docker/                    # configs da infra (localstack init, otel, prometheus, tempo, grafana)
```

A regra de dependência aponta **para dentro**: `adapter → application → domain`. O domínio não conhece HTTP nem DynamoDB; adapters implementam as *ports* definidas no domínio. Trocar de persistência (ex. DynamoDB → Postgres) só afeta o adapter outbound.

### Modelagem DynamoDB (single-table design)

Tabela única (`boilerplate`) com chaves genéricas `PK`/`SK` e um índice global `GSI1` (`GSI1PK`/`GSI1SK`):

| Item            | PK              | SK              | GSI1PK | GSI1SK              |
|-----------------|-----------------|-----------------|--------|--------------------|
| User            | `USER#<id>`     | `USER#<id>`     | `USER` | `<createdAt>#<id>` |
| Lock de e-mail  | `EMAIL#<email>` | `EMAIL#<email>` | –      | –                  |

- **Unicidade de e-mail:** a criação grava o item do usuário e o lock de e-mail numa única `TransactWriteItems` com `attribute_not_exists`, garantindo unicidade atômica.
- **Listagem:** `Query` no `GSI1` (`GSI1PK = USER`), ordenada por data, com **paginação por cursor** (`next_cursor`).

## Como rodar

### 1. Pré-requisitos
- Go 1.23+
- Docker + Docker Compose

### 2. Subir a infraestrutura
```bash
make infra-up        # localstack(dynamodb) + otel-collector + prometheus + tempo + grafana
```
O LocalStack cria a tabela automaticamente via script de init (`docker/localstack/init`).
Confira com `make db-tables`.

### 3. Configurar o ambiente
```bash
cp .env.example .env
```

### 4. Rodar a API (pela IDE ou terminal)
```bash
make run             # go run ./cmd/api
```

A API sobe em `http://127.0.0.1:8080`.

## Endpoints

| Método | Rota                  | Descrição                          |
|--------|-----------------------|------------------------------------|
| GET    | `/healthz`            | Liveness probe                     |
| GET    | `/readyz`             | Readiness (checa o DynamoDB)       |
| POST   | `/api/v1/users`       | Cria usuário                       |
| GET    | `/api/v1/users`       | Lista (`?limit=&cursor=`)          |
| GET    | `/api/v1/users/{id}`  | Busca por id                       |
| PUT    | `/api/v1/users/{id}`  | Atualiza nome                      |
| DELETE | `/api/v1/users/{id}`  | Remove                             |

Exemplo:
```bash
curl -X POST 127.0.0.1:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","email":"ada@example.com"}'

# Listagem paginada
curl "127.0.0.1:8080/api/v1/users?limit=10"
# => {"items":[...],"next_cursor":"..."}  (use o cursor na próxima página)
```

## Serviços (UI / observabilidade)

> No Windows/macOS use **`127.0.0.1`**, não `localhost`: o `localhost` resolve para IPv6 (`::1`) e as portas publicadas do Docker Desktop só respondem em IPv4, causando timeout no navegador.

| Serviço         | URL                       | Login         |
|-----------------|---------------------------|---------------|
| Grafana         | http://127.0.0.1:3000     | admin / admin |
| Prometheus      | http://127.0.0.1:9090     | –             |
| Tempo (API)     | http://127.0.0.1:3200     | –             |
| DynamoDB Admin  | http://127.0.0.1:8001     | –             |
| LocalStack      | http://127.0.0.1:4566     | –             |

A aplicação exporta traces e métricas via OTLP (`127.0.0.1:4317`) para o OTEL Collector, que distribui para Tempo (traces) e Prometheus (métricas). O Grafana já vem com os datasources provisionados e um dashboard de HTTP. O **DynamoDB Admin** permite navegar pelas tabelas do DynamoDB no LocalStack.

## Comandos úteis

```bash
make help            # lista todos os alvos
make db-tables       # lista as tabelas no DynamoDB (LocalStack)
make db-scan         # scan da tabela (debug)
make test            # testes com -race e cobertura
make lint            # golangci-lint
make infra-down      # derruba a infra
```

## Adicionando uma nova feature

1. Modele o domínio em `internal/domain/<feature>` (entidade + port).
2. Escreva os use cases em `internal/application/<feature>`.
3. Implemente o repositório em `internal/adapter/dynamodb` (defina o padrão de chaves no single-table).
4. Exponha via handlers em `internal/adapter/http` e monte no `router.go`.
5. Se precisar de novas tabelas/índices, ajuste o init em `docker/localstack/init`.
