.DEFAULT_GOAL := help

# ─── Config ──────────────────────────────────────────────────
APP            := boilerplate-api
MAIN           := ./cmd/api
DYNAMODB_TABLE ?= boilerplate
AWS_ENDPOINT   ?= http://localhost:4566

# ─── Help ────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ─── App (runs locally / from the IDE) ───────────────────────
.PHONY: run
run: ## Run the API locally
	go run $(MAIN)

.PHONY: build
build: ## Build the API binary into ./bin
	go build -o bin/$(APP) $(MAIN)

.PHONY: tidy
tidy: ## Sync go.mod/go.sum
	go mod tidy

# ─── Quality ─────────────────────────────────────────────────
.PHONY: test
test: ## Run tests with race detector and coverage
	go test -race -cover ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint installed)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format the codebase
	gofmt -w .

# ─── Infrastructure (Docker = infra only) ────────────────────
.PHONY: infra-up
infra-up: ## Start LocalStack (DynamoDB) + observability stack
	docker compose up -d

.PHONY: infra-down
infra-down: ## Stop the infrastructure
	docker compose down

.PHONY: infra-clean
infra-clean: ## Stop infra and delete volumes
	docker compose down -v

.PHONY: logs
logs: ## Tail infrastructure logs
	docker compose logs -f

# ─── DynamoDB helpers (LocalStack) ───────────────────────────
.PHONY: db-tables
db-tables: ## List DynamoDB tables in LocalStack
	docker compose exec localstack awslocal dynamodb list-tables

.PHONY: db-scan
db-scan: ## Scan the table (debug only)
	docker compose exec localstack awslocal dynamodb scan --table-name $(DYNAMODB_TABLE)

.PHONY: db-recreate
db-recreate: ## Re-run the table init script inside LocalStack
	docker compose exec localstack bash /etc/localstack/init/ready.d/01-create-tables.sh
