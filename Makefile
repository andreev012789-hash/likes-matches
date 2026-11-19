DATABASE_URL ?= postgres://likes:likes@localhost:5455/likes?sslmode=disable
HTTP_ADDR ?= :18080

.PHONY: up down run migrate-up migrate-down migrate-status

up:
	docker compose up -d

down:
	docker compose down

run:
	DATABASE_URL='$(DATABASE_URL)' HTTP_ADDR='$(HTTP_ADDR)' go run ./cmd/api

migrate-up:
	goose -dir migrations postgres '$(DATABASE_URL)' up

migrate-down:
	goose -dir migrations postgres '$(DATABASE_URL)' down

migrate-status:
	goose -dir migrations postgres '$(DATABASE_URL)' status