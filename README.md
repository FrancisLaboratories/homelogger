<p align="center">
	<img src="client/public/logoname.png" alt="HomeLogger wordmark" width="320" />
</p>

# HomeLogger

HomeLogger is a simple home maintenance and asset tracker for homeowners, inspired by the similarly-named [LubeLogger](https://github.com/hargata/lubelog). It centralizes appliances, repairs, and maintenance tasks so you can record work, receipts, and schedules for things around your home.

> [!WARNING]
> HomeLogger is still in version 0.x.x and is subject to major changes from version to version. I am developing the core features and collecting feedbacks. Expect bugs! Please open issues or feature requests

This project is in it's early stages. Expect changes. You are encouraged to [contribute](#Contributing) as well. Be mindful, this is a side project, not my full-time job, so development will be slow and incremental. The goal is to build a useful tool for myself and others. Constructive feedback is always welcome.

There is a demo available at [homelogger-demo.francislaboratories.com](https://homelogger-demo.francislaboratories.com)

This repository contains a Next.js React client and a Go (Fiber + GORM) server with a SQLite or PostgreSQL database. The project is early-stage but includes a working client and server and a small REST API defined in [server/openapi.yaml](server/openapi.yaml).

**Contents**

- **Client:** web UI built with Next.js and React ([client](client/))
- **Server:** Go API server using Fiber and GORM ([server](server/))
- **Database:** SQLite (default) or PostgreSQL

**Goals**

- Track appliances, repairs and maintenance history
- Attach files (receipts/photos) to records
- Provide a simple, local-first experience with optional Docker support

**Tech Stack**

- Client: Node, React, Bootstrap built with Vite
- Server: Go,Go Fiber web framework, GORM ORM
- Database: SQLite (default) or PostgreSQL

**Repository Layout (high level)**

- [client](client/) — Next.js app and frontend components
- [server](server/) — Go server, internal packages, and OpenAPI spec
  - [server/cmd/server/main.go](server/cmd/server/main.go)
  - [server/openapi.yaml](server/openapi.yaml)
  - [server/internal/models](server/internal/models)

## Getting started

Prerequisites

- Go (>= 1.20 recommended) for running the server locally
- Node.js (24+) and npm for the client
- Docker & Docker Compose (optional, for containerized runs)

## Docker Compose (recommended for quick start)

This repo includes an example `docker-compose.yml` for running both services together.

Start both services with:

```bash
docker compose up
```

The client will be available at http://localhost:3005

Stop and remove containers with:

```bash
docker compose down
```

## Local development

1. Start the API server

```bash
cd server
go run ./cmd/server
```

By default, server uses SQLite file under `server/data/db`. On first run it creates DB file and tables via GORM migrations in code.

To use PostgreSQL, set `DB_DIALECT=postgres` and provide connection settings (see Environment configuration below).

2. Start the client

```bash
cd client
npm install
npm run dev
```

Open http://localhost:3000 to view the Next.js app. The client expects the API to be running at the default address configured in the client environment (see `client/.env` or client code for API URL locations).

## Environment configuration

Server environment variables (create `.env` at `server/` if needed)

- `PORT` — port to run API (default `3005`)
- `DB_DIALECT` — `sqlite` (default) or `postgres`
- `DB_DIALECT` is locked on first successful server start and reused on later starts
- `DATABASE_URL` —
  - for SQLite: optional DB file path (used after `DEMO_DB_PATH`)
  - for PostgreSQL: full DSN/URL (preferred)
- `DB_DIALECT_LOCK_PATH` — optional lock file path (default `./data/db/.db_dialect`)
- `FORCE_DB_DIALECT_CHANGE` — set `true`/`1` to intentionally override existing dialect lock

SQLite-related variables

- `DEMO_DB_PATH` — overrides SQLite DB file path (used for demo mode)

PostgreSQL variables (used when `DB_DIALECT=postgres` and `DATABASE_URL` not set)

- `DB_HOST` — PostgreSQL host (default `localhost`)
- `DB_PORT` — PostgreSQL port (default `5432`)
- `DB_USER` — PostgreSQL user
- `DB_PASSWORD` — PostgreSQL password
- `DB_NAME` — PostgreSQL database name
- `DB_SSLMODE` — PostgreSQL SSL mode (default `disable`)

Demo mode variables

- `DEMO_MODE` — enable demo seed/reset mode (`true` or `1`)
- `DEMO_FILE_PATH` — optional path to demo seed JSON

Note: `DEMO_MODE` is SQLite-only. If enabled with PostgreSQL, server logs warning and disables demo mode.

Client environment variables

- The client uses Next.js environment patterns if required (check `client/next.config.js` or `client/.env.local`)

## API and docs

The server exposes a REST API. The OpenAPI spec is available at [server/openapi.yaml](server/openapi.yaml). Use it to generate clients, inspect endpoints, or run API docs tools (Swagger UI / Redoc).

## Data and uploads

- SQLite DB file is stored under [server/data/db](server/data/db)
- Uploaded files are stored under [server/data/uploads](server/data/uploads)

## Backup & export

- The app includes a server endpoint and a client settings page to download a full backup.
- The backup endpoint: `GET /backup/download` on API server. It streams ZIP containing:
  - database backup artifact
  - uploads directory with all files

The backup is a ZIP containing `data.json` (all database records) and an `uploads/` directory. Works identically for both SQLite and PostgreSQL — no dialect-specific tooling required.

## Import

The app includes a server endpoint and client settings page to import a backup ZIP. Import performs a wipe-and-replace: all existing data and uploaded files are deleted and replaced with backup contents.

- REST endpoint: `POST /backup/import` (multipart form, field name `backup`)
- Web UI: open Settings → "Import Backup" → select `.zip` file → confirm overwrite


Notes & safety

- Import is destructive. Always keep an additional copy of the original backup before proceeding.
- Restores can fail if versions mismatch; ensure server code and DB schema are compatible with backup payload version.
- The import and export endpoints are unauthenticated in this version — if you expose the server to untrusted networks, add authentication or restrict access.

## Development tips

- When changing server models, GORM auto-migrations will apply on startup (see `server/internal/database/gorm.go`).

## Contributing

Contributions are welcome. Please read the following guidelines before submitting a PR:

- Contributors Policy [CONTRIBUTORS.md](CONTRIBUTORS.md)
- AI Policy [AI_POLICY.md](AI_POLICY.md)
- Code of Conduct [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## License

This project is available under the terms of the MIT license, as shown in the [LICENSE](LICENSE) file in this repository.

## Contact

For questions or feedback, open an issue or discussion post in this repo. To privately report security vulnerabilities, please follow the guidelines in the [SECURITY.md](SECURITY.md) document.

## Further work / Roadmap

Development is ongoing. Planned features include:

- Planning/scheduling tools for seasonal maintenance and repairs
- Report creation and export functionality
- Much more to come. Also open to suggestions and contributions!
  - Please don't submit a PR for a major feature without discussing it first via an issue or discussion post.
