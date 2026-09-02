# API endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | no | Process health |
| GET | `/api/v1/health` | no | Process health |
| POST | `/api/v1/auth/login` | no | Sign in; body `{ "username", "password" }` |
| GET | `/api/v1/jobs` | yes | Cron schedule plus recent runs |
| POST | `/api/v1/jobs/run` | yes | Run scrape now |
| GET | `/api/v1/expenses` | yes | Stored shop, item, expense rows |
| GET | `/api/v1/sources` | yes | Website source placeholders |
| POST | `/api/v1/sources` | yes | Create source `{ "name", "websiteUrl", "isEnabled" }` |
| PATCH | `/api/v1/sources/:id` | yes | Update source |
| DELETE | `/api/v1/sources/:id` | yes | Delete source |
| GET | `/api/v1/forwarding` | yes | Destination POST URL and last delivery |
| PUT | `/api/v1/forwarding` | yes | Set `{ "targetApiUrl" }` |
