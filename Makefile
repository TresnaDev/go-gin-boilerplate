.PHONY: run build tidy up down logs restart

ps:
	docker compose ps

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f api

restart:
	docker compose restart api
