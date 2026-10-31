golnk
=====

A Tiny URL service built with Go's standard library plus `spf13/pflag`.

- Short URLs redirect to their original URL with HTTP 302
- Fetches and stores the `<title>` of the original site at registration time
- Honors proxy environment variables (`HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY`)
- Data is persisted to a JSON file (loaded into memory at startup)

Usage
-----

```bash
go run ./cmd/golnk
```

The server listens on `:8080` by default.

Options
-------

POSIX-compatible (short and long forms). Environment variables take precedence over the default value.

| Short | Long | Env var | Default | Description |
|-------|------|---------|---------|-------------|
| `-a` | `--addr` | `GOLNK_ADDR` | `:8080` | Listen address |
| `-f` | `--file` | `GOLNK_FILE` | `data.json` | Data file path |

API
---

**Create a short URL**

```bash
curl -X POST http://localhost:8080/api/shorten -d 'url=https://example.com/'
```

The Content-Type is `application/x-www-form-urlencoded`.

Response (the title is fetched from the original site; empty if it cannot be retrieved):

```json
{"short_url":"http://localhost:8080/c8psI6","title":"Example Domain"}
```

**Redirect**

```bash
curl -L http://localhost:8080/c8psI6   # 302 redirect to the original URL
```

An unknown ID returns 404.

Data file
---------

`data.json` schema:

```json
{
  "c8psI6": {
    "url": "https://example.com/",
    "title": "Example Domain",
    "created_at": "2026-09-08T22:49:18.058672+09:00"
  }
}
```

IDs are 6 characters (`a-zA-Z0-9`) generated with `crypto/rand`. Writes use a temp file + `rename` for atomicity.

Docker
------

```bash
docker compose up -d --build
```

Data is stored in the `golnk-data` volume. If you use a proxy, uncomment and set the environment variables shown in the compose file.

Development
-----------

Architecture
------------

Layers are separated by dependency direction, following clean architecture.

```
cmd/golnk/main.go                  # composition root (flag parsing, dependency injection, server startup)
internal/domain/                   # Link entity, domain errors (no external dependencies)
internal/application/              # Shortener use case + ports (interfaces)
internal/infrastructure/
  repo/                            # JSON file persistence (LinkRepository implementation)
  fetcher/                         # <title> fetching (TitleFetcher implementation)
  idgen/                           # crypto/rand ID generation (IDGenerator implementation)
internal/interface/http/           # HTTP handlers (form parsing, responses, error mapping)
```

Dependencies point inward (domain → application → external). `application` abstracts external dependencies behind ports, so unit tests can use fakes.

Tests
-----

```bash
go test ./...
```

- `internal/application`: unit tests for the use case (fake repository / fetcher / generator)
- `internal/infrastructure`: persistence and title-fetching tests
- `internal/interface/http`: handler tests using `httptest`