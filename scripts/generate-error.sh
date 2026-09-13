#!/usr/bin/env bash
# Forces a genuine 500 (not a validation/404) by briefly stopping postgres,
# hitting the API, then restarting postgres. Only use against a local demo
# stack - it's a few seconds of real DB downtime.
set -euo pipefail

API="${API_URL:-http://localhost:8080}"

echo "Stopping postgres..."
docker compose --env-file rest.env stop postgres

echo "Hitting API while DB is down (expect 500s)..."
for _ in $(seq 1 5); do
  curl -s -o /dev/null -w "status=%{http_code}\n" "$API/api/v1/tasks/"
  sleep 1
done

echo "Restarting postgres..."
docker compose --env-file rest.env start postgres

echo "Done. Give the app a couple seconds to reconnect, then traffic should be healthy again."
