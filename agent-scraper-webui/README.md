# agent-scraper-webui

Vue + Tailwind + Shadcn + Inter + PWA dashboard for Agent Scraper.

## Run

1. Start sibling `agent-scraper-api` on port **8196**
2. Copy `.env.example` → `.env` if needed (`VITE_API_BASE_URL` may stay empty; Vite proxies `/api`)
3. `npm install`
4. `npm run dev` — WebUI at http://127.0.0.1:5196/jobs

Default login: `armin` / `dopadopa123`

## Pages

See [docs/pages.md](docs/pages.md).
