<p align="center">
 <img src="client/public/logoname.png" alt="HomeLogger wordmark" width="320" />
</p>

# HomeLogger

HomeLogger is a simple home maintenance and asset tracker for homeowners, inspired by the similarly-named [LubeLogger](https://github.com/hargata/lubelog). It centralizes appliances, repairs, and maintenance tasks so you can record work, receipts, and schedules for things around your home.

> [!WARNING]
> HomeLogger is still in version 0.x.x and is subject to major changes from version to version. I am developing the core features and collecting feedbacks. Expect bugs! Please open issues or feature requests

This project is in it's early stages. Expect changes. You are encouraged to [contribute](#contributing) as well. Be mindful, this is a side project, not my full-time job, so development will be slow and incremental. The goal is to build a useful tool for myself and others. Constructive feedback is always welcome.

There is a demo available at [homelogger-demo.francislaboratories.com](https://homelogger-demo.francislaboratories.com)

This repository contains a Go (Fiber + GORM) server with a built-in React SPA (served via Fiber static middleware) and a SQLite or PostgreSQL database. The production Docker image is a single monolith container. For development, the client and server can run separately. The project is early-stage but includes a working UI and a small REST API defined in [server/openapi.yaml](server/openapi.yaml).

**Contents**

- **Client:** web UI built with Vite and React ([client](client/))
- **Server:** Go API server using Fiber and GORM ([server](server/))
- **Database:** SQLite (default) or PostgreSQL

**Goals**

- Track appliances, repairs and maintenance history
- Attach files (receipts/photos) to records
- Provide a simple, local-first experience with optional Docker support

**Tech Stack**

- Client: Node, React, Bootstrap built with Vite
- Server: Go, Fiber web framework, GORM ORM
- Database: SQLite (default) or PostgreSQL

**Repository Layout (high level)**

- [client](client/) — React app and frontend components
- [server](server/) — Go server, internal packages, and OpenAPI spec
  - [server/cmd/server/main.go](server/cmd/server/main.go)
  - [server/openapi.yaml](server/openapi.yaml)
  - [server/internal/models](server/internal/models) — data models
  - [server/internal/database](server/internal/database) — GORM setup, migrations, backup/import
  - [server/internal/demo](server/internal/demo) — demo mode seed/reset logic
  - [server/internal/version](server/internal/version) — build version info
- [docker/](docker/) — alternate Docker Compose configurations (dev, demo, postgres)

## Getting started

You have two options: run everything with Docker (recommended, no setup needed) or run locally if you want to make changes to the code.

### Option 1: Docker (easiest)

You will need [Docker](https://docker.com) installed on your machine. Create a docker-compose file with the following contents:

```yaml
services:
  homelogger:
    image: francislaboratories/homelogger:latest
    container_name: homelogger
    ports:
      - "3005:3005"
    volumes:
      - homelogger_data:/app/data
    restart: unless-stopped
volumes:
  homelogger_data:
```

Then, run the following command in the same directory as your `docker-compose.yml` file:

```bash
docker compose up
```

Open <http://localhost:3005> in your browser. The app will be ready to use — it creates and manages its own database file automatically.

To stop, press `Ctrl+C` or run:

```bash
docker compose down
```

There are also a few alternative setups in the [docker/](docker/) folder if you need something different:

| File | What it does |
|------|--------------|
| `dev.docker-compose.yml` | Lets you work on the code with the app running in containers |
| `postgres.docker-compose.yml` | Uses a different database system (PostgreSQL) instead of the default file-based one |
| `demo.docker-compose.yml` | Starts the app with sample data pre-loaded so you can explore |
| `combo.docker-compose.yml` | Builds the app from the source code instead of using a pre-made image |

### Option 2: Run locally (for developers)

You'll need to have the following installed:

- **Go** (version 1.25 or newer) for the server
- **Node.js** (version 24 or newer, comes with npm) for the web interface

1. **Set up environment variables**

   Copy the client example file and adjust if needed:

   ```bash
   cp client/.env.example client/.env.local
   ```

   The defaults work out of the box — you only need to change them if you want to use a different database or port. See the [Environment configuration](#environment-configuration) section for all available options.

   > Server settings are set through environment variables in your terminal. For example, to run on a different port:
   > ```bash
   > export PORT=4000
   > ```
   > You can also prefix them inline: `PORT=4000 go run ./cmd/server`

2. **Start the server**

   Open a terminal, go to the `server` folder, and run:

   ```bash
   cd server
   go run ./cmd/server
   ```

   The server uses a file-based database by default — it will create everything it needs on the first run.

3. **Start the web interface**

   Open another terminal, go to the `client` folder, and run:

   ```bash
   cd client
   npm install
   npm run dev
   ```

4. Open http://localhost:5173 in your browser.

## Environment configuration

Create `.env` at `server/` for server vars. Create `client/.env.local` for client vars.

`DB_DIALECT` is locked on first successful start and reused on later starts.

> [!CAUTION]
> `FORCE_DB_DIALECT_CHANGE` overrides the dialect lock. Only use this if you understand the consequences — switching dialects after data exists will not migrate your data.

**Server variables**

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `PORT` | `3005` | No | API server listen port |
| `DB_DIALECT` | `sqlite` | No | `sqlite` or `postgres` |
| `DATABASE_URL` | — | Conditional | SQLite: path to DB file (optional). Postgres: full connection string like `postgres://user:pass@host:5432/dbname` (preferred over individual `DB_*` vars) |
| `DB_DIALECT_LOCK_PATH` | `./data/db/.db_dialect` | No | Dialect lock file path |
| `FORCE_DB_DIALECT_CHANGE` | — | No | Override dialect lock (advanced — see caution above). Set to `true` or `1` |
| `DEMO_DB_PATH` | — | No | Override SQLite DB file path |
| `DEMO_MODE` | — | No | Enable demo seed/reset. Set to `true` or `1`. SQLite only |
| `DEMO_FILE_PATH` | — | No | Path to demo seed JSON |
| `DB_HOST` | `localhost` | Conditional | Postgres host (when `DB_DIALECT=postgres` and no `DATABASE_URL`) |
| `DB_PORT` | `5432` | Conditional | Postgres port |
| `DB_USER` | — | Conditional | Postgres user |
| `DB_PASSWORD` | — | Conditional | Postgres password |
| `DB_NAME` | — | Conditional | Postgres database name |
| `DB_SSLMODE` | `disable` | No | Postgres SSL mode |
| `LOG_CONSOLE` | `true` | No | Console request logging. Set to `true` or `false` |
| `LOG_FILE` | — | No | File path for request logs (e.g. `/var/log/homelogger.log`). Leave unset or blank to disable file logging |

**Client variables**

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_SERVER_URL` | Yes | API server URL (e.g. `http://localhost:3005` for local dev, `/api` when served via Docker monolith). Only needed when running the client standalone or building locally |

## API and docs

The server exposes a REST API. The OpenAPI spec is available at [server/openapi.yaml](server/openapi.yaml). Use it to generate clients, inspect endpoints, or run API docs tools (Swagger UI / Redoc).

## Data and uploads

- SQLite DB file is stored under [server/data/db](server/data/db)
- Uploaded files are stored under [server/data/uploads](server/data/uploads)
- Server accepts uploads up to 100 MB (configurable via `BodyLimit` in server code)
- Production Docker container uses a healthcheck (`prod.healthcheck.sh`)

## Backup & export

- The app includes a server endpoint and a client settings page to download a full backup.
- The backup endpoint: `GET /backup/download` on API server.

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
