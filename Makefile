.PHONY: up down test-core dev-portal migrate-up migrate-down migrate-status

-include .env
export

MIGRATIONS_DIR := packages/database/migrations

up:
	docker compose up -d

down:
	docker compose down

test-core:
	cd services/core && go test ./...

dev-portal:
	pnpm install && pnpm --workspace apps/portal run dev

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status
