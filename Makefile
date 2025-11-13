# === CONFIG ===
ENV_FILE := .env
-include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE))


.PHONY: help up down gen test db/migrate db/reset run

help:
	@echo "Available targets:"
	@echo "  up            Start local Postgres via Docker"
	@echo "  down          Stop local Postgres container"
	@echo "  gen           Regenerate SQLC code"
	@echo "  db/migrate    Apply DB migrations"
	@echo "  db/reset      Drop and re-run migrations"
	@echo "  test          Run all Go tests"
	@echo "  run           Launch Dugout app"

up:
	docker compose up -d db

down:
	docker compose down

gen:
	sqlc generate

db/migrate:
	goose -dir migrations postgres "$(DB_URL)" up

db/reset:
	goose -dir migrations postgres "$(DB_URL)" reset
	goose -dir migrations postgres "$(DB_URL)" up

test:
	go test ./... -race -cover

run:
	go run ./cmd/dugout/
