# Sharing Vision — Backend

Golang blog/article microservice for the PT. Sharing Vision hiring test.
Runs on Cloudflare-backed infrastructure (Docker Compose + Cloudflare Tunnel) but is plain Go + MySQL underneath.

## Endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/article/` | Create article `{title, content, category, status?}` |
| GET | `/api/article/` | List all (`?status=publish\|draft\|thrash`) |
| GET | `/api/article/{id}` | Get one article |
| PUT | `/api/article/{id}` | Full update incl. `status` |
| DELETE | `/api/article/{id}` | Trash (soft delete → `status=thrash`) |
| GET | `/api/article/preview` | Published list `?page=&per_page=` (paginated) |
| GET | `/api/health` | Liveness |

`status` is one of `publish`, `draft`, `thrash` (default `draft`).

## Database

Table `posts` (exact spec):

```sql
CREATE TABLE posts (
  `Id`           INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `Title`        VARCHAR(200) NOT NULL,
  `Content`      TEXT,
  `Category`     VARCHAR(100),
  `Created_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `Updated_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `Status`       VARCHAR(100) DEFAULT 'draft' CHECK (Status IN ('publish','draft','thrash'))
);
```

Two ways to create it:

1. **Manual (task, 80 pts):** `mysql -u root -p article < sql/manual_create.sql`
2. **Tool (task, 20 pts):** embedded [golang-migrate](https://github.com/golang-migrate/migrate)
   migrations in `migrations/`, applied automatically at startup.

## Run locally

```bash
# 1. Start a MySQL 8 and create database + user:
docker run -d --name sv-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=article \
  -e MYSQL_USER=sv -e MYSQL_PASSWORD=sv -p 3306:3306 mysql:8

# 2. Run the API (config from env, defaults in .env.example):
cp .env.example .env   # adjust MYSQL_PASSWORD etc.
go run ./cmd/api
```

```bash
curl -X POST localhost:8080/api/article/ -H 'Content-Type: application/json' \
  -d '{"Title":"Hello","Content":"World","Category":"News","Status":"publish"}'
```

## Test

```bash
go test ./...
```

## Deploy (CI)

Push to `main` → GitHub Actions builds the image, pushes to `ghcr.io/towgenik/sharing-vision-backend`, then SSHes into the LXC over Tailscale and runs `docker compose up -d api`.

