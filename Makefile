.PHONY: up down test-core dev-portal

up:
	docker compose up -d

down:
	docker compose down

test-core:
	cd services/core && go test ./...

dev-portal:
	npm install && npm --workspace apps/portal run dev
