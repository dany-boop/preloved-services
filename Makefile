# ============================================================
# Preloved services — Makefile
# Run: make help  to see all commands
# ============================================================

# Load environment variables from .env if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Defaults if not defined in .env
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= Drew2424
POSTGRES_DB ?= preloved
POSTGRES_PORT ?= 5432
MONGO_USER ?= aguser
MONGO_PASSWORD ?= change_me_in_production
REDIS_PASSWORD ?= change_me_in_production

# Database Connection Strings
DB_DOCKER_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@preloved_postgres:5432/$(POSTGRES_DB)?sslmode=disable
DB_LOCAL_URL  := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

# Configurations
MIGRATIONS_PATH := infra/migrations
MIGRATE_IMAGE   := migrate/migrate:v4.17.0
MIGRATE         := docker run --rm -v "$(CURDIR)/$(MIGRATIONS_PATH):/migrations" --network preloved_net $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_DOCKER_URL)"

.PHONY: help setup dev-infra dev-infra-down dev-infra-clean logs \
        migrate-up migrate-down migrate-reset migrate-version migrate-force migration-create \
        postgres-cli redis-cli mongo-cli \
        tidy lint test test-coverage \
        run-auth run-user run-chat run-notification run-search run-ai run-media

# Colors for output
GREEN  := \033[32m
YELLOW := \033[33m
RESET  := \033[0m

help: ## Show this help
	@echo "$(GREEN)Preloved services$(RESET)"
	@echo "-------------------"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

# ── Infrastructure ──────────────────────────────

dev-infra: ## Start all active infrastructure containers
	@echo "$(GREEN)Starting active infrastructure...$(RESET)"
	docker compose up -d
	@echo "$(GREEN)✅ Infrastructure running$(RESET)"

dev-infra-down: ## Stop all infrastructure containers
	docker compose down

dev-infra-clean: ## Stop containers and delete all data volumes (destructive)
	docker compose down -v
	@echo "$(YELLOW)⚠️  All data volumes removed$(RESET)"

logs: ## Tail container logs: make logs s=<service_name> (default all)
	docker compose logs -f $(s)

# ── Database & Migrations ───────────────────────

migration-create: ## Create new migration files: make migration-create name=<migration_name>
ifndef name
	$(error Error: migration name 'name' is required. Example: make migration-create name=create_users_table)
endif
	@echo "Creating migration files for '$(name)'..."
	@docker run --rm -v "$(CURDIR)/$(MIGRATIONS_PATH):/migrations" $(MIGRATE_IMAGE) create -ext sql -dir /migrations -seq $(name)
	@echo "$(GREEN)✅ Migration files created in $(MIGRATIONS_PATH)/$(RESET)"

migrate-up: ## Run all pending migrations
	@echo "$(GREEN)Running migrations up...$(RESET)"
	@$(MIGRATE) up
	@echo "$(GREEN)✅ Migrations applied successfully$(RESET)"

migrate-down: ## Rollback migrations (default N=1): make migrate-down n=2
	@echo "$(GREEN)Rolling back migrations...$(RESET)"
	@$(MIGRATE) down $(or $(n),1)
	@echo "$(GREEN)✅ Rollback completed$(RESET)"

migrate-reset: ## Wipe database and run all migrations from scratch
	@echo "$(GREEN)Wiping database and running all migrations...$(RESET)"
	@$(MIGRATE) drop -f
	@$(MIGRATE) up
	@echo "$(GREEN)✅ Database reset complete$(RESET)"

migrate-version: ## Print the current database migration version
	@$(MIGRATE) version

migrate-force: ## Force migration version to V (recovers from dirty-state): make migrate-force v=<version>
ifndef v
	$(error Error: version 'v' is required. Example: make migrate-force v=4)
endif
	@echo "$(YELLOW)Forcing migration version to $(v)...$(RESET)"
	@$(MIGRATE) force $(v)
	@echo "$(GREEN)✅ Migration version forced$(RESET)"

# ── Shell Helpers ───────────────────────────────

postgres-cli: ## Open PostgreSQL psql interactive shell
	docker exec -it preloved_postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

redis-cli: ## Open Redis interactive shell (when enabled)
	docker exec -it preloved_redis redis-cli -a $(REDIS_PASSWORD)

mongo-cli: ## Open MongoDB interactive shell (when enabled)
	docker exec -it preloved_mongodb mongosh -u $(MONGO_USER) -p $(MONGO_PASSWORD) --authenticationDatabase admin

# ── Go Development ──────────────────────────────

tidy: ## Run go mod tidy in all services and shared
	@echo "$(GREEN)Tidying Go modules...$(RESET)"
	cd shared && go mod tidy
	cd services/auth-service && go mod tidy
	cd services/user-service && go mod tidy
	cd services/chat-service && go mod tidy
	cd services/notification-service && go mod tidy
	cd services/search-service && go mod tidy
	cd services/ai-service && go mod tidy
	cd services/media-service && go mod tidy
	go work sync
	@echo "$(GREEN)✅ All modules tidied$(RESET)"

lint: ## Run golangci-lint on all services
	@which golangci-lint || (echo "Install: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

test: ## Run all tests
	@echo "$(GREEN)Running tests...$(RESET)"
	go test ./... -v -timeout 30s

test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(RESET)"

# ── Service Dev Servers ─────────────────────────

run-auth: ## Run auth-service locally
	cd services/auth-service && go run cmd/main.go

run-user: ## Run user-service locally
	cd services/user-service && go run cmd/main.go

run-chat: ## Run chat-service locally
	cd services/chat-service && go run cmd/main.go

run-notification: ## Run notification-service locally
	cd services/notification-service && go run cmd/main.go

run-search: ## Run search-service locally
	cd services/search-service && go run cmd/main.go

run-ai: ## Run ai-service locally
	cd services/ai-service && go run cmd/main.go

run-media: ## Run media-service locally
	cd services/media-service && go run cmd/main.go

# ── Setup ───────────────────────────────────────

setup: ## First-time project setup
	@echo "$(GREEN)Setting up Preloved services...$(RESET)"
ifeq (,$(wildcard .env))
	@copy .env.example .env || cp .env.example .env
	@echo "Created .env from .env.example"
endif
	@$(MAKE) dev-infra
	@echo "Waiting for PostgreSQL to be healthy..."
	@docker exec preloved_postgres sh -c "until pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB); do sleep 1; done"
	@echo "Running migrations..."
	@$(MAKE) migrate-up
	@echo "$(GREEN)✅ Setup complete! Run: make run-auth$(RESET)"
