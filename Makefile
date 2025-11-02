# === CONFIG ===
ENV_FILE := .ENV
INCLUDE $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE))

DB_URL ?= postgres://postgres:postgres@localhost:5432/roster?sslmode=disable

.PHONY: help up down gen test db/migrate db/reset

help:
	@echo "Available targets:"
	@echo "  up            Start local Postgres via Docker"
	@echo "  down          Stop local Postgres container"
	@echo "  gen           Regenerate SQLC code"
	@echo "  db/migrate    Apply DB migrations"
	@echo "  db/rest       Drop and re-run migrations"
	@echo "  test          Run all Go tests"

up:
	docker compose up -d db

down:
	docker compose down

gen:
	sqlc generate

db/migrate:
	goose -dir migrations postgres "$(DB_URL)" up

db/rest:
	goose -dir migrations postgres "$(DB_URL)" reset
	goose -dir migrations postgres "$(DB_URL)" up

test:
	go test ./...
