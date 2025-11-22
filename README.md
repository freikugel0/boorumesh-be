# BooruMesh

BooruMesh is a backend service that **aggregates image board APIs** (Danbooru, Konachan, etc.) into a **unified image schema**.  
The goal: developers can configure multiple sources via a dev API, and clients can consume a single, consistent API.

## Features

- **Dev Source Management**
  - Add new image board sources via API (`/dev/sources`)
  - Store source configuration in PostgreSQL (Neon):
    - `base_url`
    - request configuration (path, query params, headers)
    - field mapping from upstream JSON → unified `Image` schema
    - defaults (max limit, timeout, etc.)
  - Enable/disable sources (planned via PATCH)

- **Source Fetch API**
  - `GET /api/:source?tags=...&page=...&limit=...`
  - Fetch from upstream image board API using stored source config
  - Normalize responses into a unified `Image` schema

- **Health Check**
  - `GET /health` → simple service liveness probe

Planned roadmap:

- `PATCH /dev/sources/:code` to update source config & toggle `enabled`
- Aggregation endpoint:
  - `GET /api/search?tags=...` → fan-out to all enabled sources and merge results
- Auth & rate limiting (if needed)

---

## Tech Stack

- **Language**: Go
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL (Neon)
- **DB Driver**: `pgx` (stdlib)
- **HTTP Client**: `net/http` (optionally wrapped with `resty`)

---

## Project Structure (overview)

```txt
cmd/
  api/
    main.go               # HTTP server entrypoint

internal/
  http/
    router.go             # Gin router setup
    handler/
      dev_source.go       # handlers for /dev/sources
      api_handler.go      # handler for /api/:source

  service/
    dev_source_service.go   # business logic: create/get sources
    source_fetch_service.go # business logic: fetch from upstream APIs

  repository/
    source_postgres.go    # SourceRepository Postgres implementation

  domain/
    source.go             # Source + config types
    image.go              # unified Image schema (API output)
```

---

## Prerequisites

- Go 1.21+ (recommended)
- PostgreSQL (Neon or local)
- Git

---

## Configuration

The application reads the database connection from the `DATABASE_URL` environment variable:

```bash
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=require
```

For Neon, `sslmode=require` is typically required.

The `sources` table is expected to have (at least) the following columns:

- `id` (bigserial, PK)
- `code` (text, unique)
- `name` (text)
- `base_url` (text)
- `enabled` (boolean)
- `request` (jsonb)
- `mapping` (jsonb)
- `defaults` (jsonb)
- `created_at` (timestamptz)
- `updated_at` (timestamptz)

(See migrations/schema files in this repo for the exact definition.)

---

## Running Locally

1. Clone the repository:

```bash
git clone https://github.com/your-username/boorumesh-be.git
cd boorumesh-be
```

2. Export environment variables:

```bash
export DATABASE_URL="postgres://user:password@host:port/dbname?sslmode=require"
```

3. Run the API server:

```bash
go run ./cmd/api
```

By default the server listens on `:8080` (check `main.go` if you change it).

---

## API Overview

### Health Check

```http
GET /health
```

Response:

```json
{ "status": "ok" }
```

---

### Dev: Create Source

```http
POST /dev/sources
Content-Type: application/json
```

Example body (Danbooru-like):

```json
{
  "code": "danbooru",
  "name": "Danbooru",
  "base_url": "https://danbooru.donmai.us",
  "request": {
    "posts_path": "/posts.json",
    "tags_param": "tags",
    "limit_param": "limit",
    "page_param": "page",
    "headers": {
      "User-Agent": "boorumesh/1.0"
    }
  },
  "mapping": {
    "fields": {
      "id":          { "key": "id" },
      "file_url":    { "key": "file_url" },
      "preview_url": { "key": "preview_file_url" },
      "sample_url":  { "key": "large_file_url" },
      "rating":      { "key": "rating" },
      "tags":        { "key": "tag_string" },
      "has_children":{ "key": "has_children" },
      "parent_id":   { "key": "parent_id" },
      "md5":         { "key": "md5" },
      "created_at":  { "key": "created_at" }
    }
  },
  "defaults": {
    "max_limit": 100,
    "timeout_ms": 10000
  }
}
```

---

### Dev: Get Source by Code

```http
GET /dev/sources/:code
```

Example:

```http
GET /dev/sources/danbooru
```

---

### Fetch Images by Source

```http
GET /api/:source?tags=...&page=...&limit=...
```

Example:

```http
GET /api/danbooru?tags=hakurei_reimu&limit=10
```

Response (unified `Image` schema example):

```json
[
  {
    "id": "123456",
    "source": "danbooru",
    "created_at": "2025-11-22T00:00:00Z",
    "image_src_url": "https://danbooru.donmai.us/posts/123456",
    "rating": "s",
    "tags": ["hakurei_reimu", "touhou"],
    "has_children": false,
    "parent_id": "",
    "md5": "abcdef123456...",
    "preview_url": "https://...",
    "sample_url": "https://...",
    "file_url": "https://..."
  }
]
```

---

## Development Notes

- Architecture uses a simple layered approach:
  - **handler (HTTP)** → **service (business logic)** → **repository (DB)** → **domain (models)**
- `context.Context` is passed through services and repositories for:
  - request-scoped cancellation
  - per-source timeouts (when fetching upstream APIs)

---

## License

TBD (MIT or similar).
