
<p align="center">
	<img src="client/public/logoname.png" alt="HomeLogger wordmark" width="320" />
</p>

# HomeLogger

HomeLogger is a simple home maintenance and asset tracker for homeowners. It centralizes appliances, repairs, and maintenance tasks so you can record work, receipts, and schedules for things around your home.

This repository contains a Next.js React client and a Go (Fiber + GORM) server with an SQLite database. The project is early-stage but includes a working client and server and a small REST API defined in [server/openapi.yaml](server/openapi.yaml).

**Contents**
- **Client:** web UI built with Next.js and React ([client](client/))
- **Server:** Go API server using Fiber and GORM ([server](server/))
- **Database:** SQLite database stored in [server/data/db](server/data/db)

**Goals**
- Track appliances, repairs and maintenance history
- Attach files (receipts/photos) to records
- Provide a simple, local-first experience with optional Docker support

**Tech Stack**
- Client: Node, Next.js, React, Bootstrap
- Server: Go, Fiber web framework, GORM ORM
- Database: SQLite

**Repository Layout (high level)**
- [client](client/) — Next.js app and frontend components
- [server](server/) — Go server, internal packages, and OpenAPI spec
	- [server/cmd/server/main.go](server/cmd/server/main.go)
	- [server/openapi.yaml](server/openapi.yaml)
	- [server/internal/models](server/internal/models)

Getting started
---------------

Prerequisites
 - Go (>= 1.20 recommended) for running the server locally
 - Node.js (24+) and npm for the client
 - Docker & Docker Compose (optional, for containerized runs)

 Docker Compose (recommended for quick start)
----------------------------------

This repo includes a `docker-compose.yml` for running both services together.

Start both services with:

```bash
docker compose up --build
```

The client will be available at http://localhost:3005

Stop and remove containers with:
 
```bash
docker compose down
```

 Local development
----------------------------------

1) Start the API server

```bash
cd server
go run ./cmd/server
```

The server uses an SQLite file under `server/data/db`. On first run it will create the DB and necessary tables via GORM migrations in the code.

2) Start the client

```bash
cd client
npm install
npm run dev
```

Open http://localhost:3000 to view the Next.js app. The client expects the API to be running at the default address configured in the client environment (see `client/.env` or client code for API URL locations).

Environment configuration
-------------------------

Server environment variables (create `.env` at `server/` if needed)
- `PORT` — port to run the API (default 8080)
- `DATABASE_URL` — (optional) path or DSN for SQLite; default is `server/data/db/homelogger.db` or similar used by code

Client environment variables
- The client uses Next.js environment patterns if required (check `client/next.config.js` or `client/.env.local`)

API and docs
------------

The server exposes a REST API. The OpenAPI spec is available at [server/openapi.yaml](server/openapi.yaml). Use it to generate clients, inspect endpoints, or run API docs tools (Swagger UI / Redoc).

Data and uploads
----------------

- SQLite DB file is stored under [server/data/db](server/data/db)
- Uploaded files are stored under [server/data/uploads](server/data/uploads)

Backup & export
----------------

- The app includes a server endpoint and a client settings page to download a full backup.
- The backup endpoint: `GET /backup/download` on the API server. It streams a ZIP containing:
	- the SQLite database file
	- the uploads directory with all files

Notes & safety
- Always keep an additional copy of the original DB before overwriting.
- If your server is behind Docker with volumes, restore the files into the host path used by the volume or restore directly inside the running container (use `docker cp` or mount the volume and replace files), then restart the container.
- Restores can fail if versions mismatch; ensure your server code and SQLite driver versions are compatible with the DB file.
	- the SQLite database file (under `data/db/`), and
	- the `data/uploads/` directory with all uploaded files.
- Client settings: open the web UI and go to `Settings` to use the "Download Backup" button which calls the endpoint and triggers a browser download.

Security note: the backup endpoint is unauthenticated in this version — if you expose the server to untrusted networks, add authentication or restrict access.

Development tips
----------------
- When changing server models, GORM auto-migrations will apply on startup (see `server/internal/database/gorm.go`).

Contributing
------------

Contributions are welcome. Suggested workflow:

1. Fork the repo
2. Create a feature branch
3. Open a PR with a clear description of changes

If you plan to add larger features (new API endpoints, DB schema changes) open an issue first to discuss the design.

License
-------

This project is available under the terms of the MIT license, as shown in the [LICENSE](LICENSE) file in this repository.

Contact
-------

For questions or feedback, open an issue or discussion post in this repo. 

Further work / Roadmap
----------------------
Development is ongoing. Planned features include:

- Planning tools for seasonal maintenance and repairs
- Enhanced reporting and export options

See [plan.md](plan.md) for tentative feature plans and priorities.
