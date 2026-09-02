# agent-scraper-api

Gin API that runs a daily scrape job, stores shop / item / expense rows in PostgreSQL, and POSTs them to a destination API.

Website list, exact run time, and destination URL are placeholders until configured.

## Run

1. Start Postgres: `docker compose up -d`
2. Copy `.env.example` → `.env`
3. `go run ./cmd/server`

Default login: `armin` / `dopadopa123`

Dev ports: API **8196**, Postgres **5456**

Default schedule: `0 2 * * *` (02:00 local). Change `CRON_SCHEDULE` later. Leave `TARGET_API_URL` empty until the destination API is known.

## Endpoints

See [docs/endpoints.md](docs/endpoints.md).
