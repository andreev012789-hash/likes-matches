DATABASE_URL ?= postgres://likes:likes@localhost:5455/likes?sslmode=disable
HTTP_ADDR ?= :8081

.PHONY: up down run

up:
	docker compose up -d

down:
	docker compose down

run:
	DATABASE_URL='$(DATABASE_URL)' HTTP_ADDR='$(HTTP_ADDR)' go run ./cmd/api