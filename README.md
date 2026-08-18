# Sharing Vision — Backend (Microservice)

Golang blog/article microservice for the PT. Sharing Vision hiring test
("Test Backend – Sharing Vision", Post Article use case).
Plain Go (chi) + MySQL, deployed via Docker Compose behind a Cloudflare Tunnel.

## Endpoints (spec §3 — paths are exact)

| No | Method | URL | Body | Description |
|:--|:--|:--|:--|:--|
| 1 | `POST` | `/article/` | `{title, content, category, status}` | Create a new article (`201`) |
| 2 | `GET` | `/article/{limit}/{offset}` | — | List all articles, paginated by `limit`+`offset` path params; body is a JSON array, total in `X-Total-Count` header. Optional `?status=` filter |
| 3 | `GET` | `/article/{id}` | — | Article for the requested id (`200`/`404`) |
| 4 | `PUT` or `PATCH` | `/article/{id}` | `{title, content, category, status}` | Modify the article for the requested id |
| 5 | `DELETE` | `/article/{id}` | — | Delete the article (soft delete → `status=thrash`, `204`) |

Same routes are also mounted under `/api/article/...` (what the web frontend calls
through the nginx `/api` reverse proxy). `GET /api/health` is a liveness probe.

## Validation (spec §4 — applied on create AND update)

| Field | Rule |
|---|---|
| `title` | **Required**, minimum **20** chars, max 200 |
| `content` | **Required**, minimum **200** chars |
| `category` | **Required**, minimum **3** chars, max 100 |
| `status` | **Required** — one of `publish`, `draft`, `thrash` |

Validation failures return `422` with `{"error": "..."}`.

## Database (spec §2)

| Weight | What | How |
|:--|:--|:--|
| 80 pts | `posts` table in the `article` database | **Manually** — `mysql article < sql/manual_create.sql` |
| 20 pts | the `article` database | **Migration in Golang** — the app creates it on boot (`CREATE DATABASE IF NOT EXISTS`, idempotent) |

`posts` schema (exact spec columns):

```sql
CREATE TABLE posts (
  `Id`           INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `Title`        VARCHAR(200) NOT NULL,
  `Content`      TEXT,
  `Category`     VARCHAR(100),
  `Created_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `Updated_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `Status`       VARCHAR(100) DEFAULT 'draft'
);
```

Status values stored lower-case: `publish` | `draft` | `thrash` (spec spelling).

## Postman collection (spec §5)

Import [`postman/sharing-vision.postman_collection.json`](postman/sharing-vision.postman_collection.json).
It contains a request for every endpoint plus a validation-failure demo. Set the
`baseUrl` collection variable to the API origin (default: `https://sv.farrel.moe`).

## Run locally

```bash
# 1. MySQL 8 (creates the `article` DB + `sv` user):
docker run -d --name sv-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=article \
  -e MYSQL_USER=sv -e MYSQL_PASSWORD=sv -p 3306:3306 mysql:8

# 2. Manual table task (80 pts):
mysql -u sv -psv article < sql/manual_create.sql

# 3. Run the API — its Golang migration creates the DB (20 pts, no-op if present):
cp .env.example .env   # adjust MYSQL_PASSWORD etc.
go run ./cmd/api
```

Example calls:

```bash
# create (must satisfy min-length rules)
curl -X POST localhost:8080/article/ -H 'Content-Type: application/json' -d '{
  "title":"A valid article title long enough to pass validation",
  "content":"'$(head -c 210 < /dev/zero | tr '\0' 'x')'",
  "category":"Tech",
  "status":"publish"}'

# list, page style (limit/offset)
curl "localhost:8080/article/10/0"
curl "localhost:8080/article/10/0?status=publish"

# get / update / delete
curl localhost:8080/article/1
curl -X PUT localhost:8080/article/1 -H 'Content-Type: application/json' < body.json
curl -X DELETE localhost:8080/article/1
```

## Test

```bash
go test ./...
```

## Deploy (CI)

Push to `main` → GitHub Actions tests + builds the image, ships it over SSH
(`docker save | gzip | ssh deploy@… | docker load`), then `docker compose up -d api`.
No registry, no published host ports — ingress only via the Cloudflare Tunnel.
