.PHONY: build tidy up down logs

tidy:
	go mod tidy

build:
	docker compose --env-file rest.env build --no-cache

up:
	docker compose --env-file rest.env up -d
	docker image prune -f

down:
	docker compose --env-file rest.env down

logs:
	docker compose --env-file rest.env logs -f api
