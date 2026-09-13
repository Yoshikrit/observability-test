.PHONY: build tidy up down logs traffic traffic-error

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

# Generates steady request traffic (mix of 2xx/4xx) so logs/traces/metrics
# have real data. DURATION/INTERVAL env vars override the defaults (60s/1s).
traffic:
	./scripts/generate-traffic.sh

# Forces a genuine 500 by briefly stopping postgres. Run traffic in another
# terminal first so there's a baseline to compare the error spike against.
traffic-error:
	./scripts/generate-error.sh
