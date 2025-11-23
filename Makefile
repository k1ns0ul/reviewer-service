.PHONY: help migrate-up migrate-down migrate-force db-create build run test docker-up docker-down scripts-permission

DB_DSN := postgresql://postgres:postgres@localhost:5433/reviewer_service?sslmode=disable

help:
	@echo "make migrate-up         - Apply migrations"
	@echo "make migrate-down       - Rollback migration"
	@echo "make migrate-force      - Force migration version"
	@echo "make db-create          - Create database"
	@echo "make build              - Build app"
	@echo "make run                - Run app"
	@echo "make test               - Run tests"
	@echo "make docker-up          - Start docker services"
	@echo "make docker-down        - Stop docker services"
	@echo "make scripts-permission - Make scripts executable"

migrate-up:
	migrate -path migrations -database "$(DB_DSN)" up

migrate-down:
	migrate -path migrations -database "$(DB_DSN)" down 1

migrate-force:
	migrate -path migrations -database "$(DB_DSN)" force $(VERSION)

db-create:
	./scripts/init-db.sh

build:
	go build -o bin/app cmd/api/main.go

run:
	DB_PORT=5433 go run cmd/api/main.go

test:
	go test -v ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

scripts-permission:
	chmod +x scripts/*.sh