# DML Enterprise API

This is the Go backend and PostgreSQL setup for DML, following an enterprise-grade layered architecture.

## Requirements
* [Go 1.22+](https://golang.org/doc/install)
* [OrbStack](https://orbstack.dev/) (or Docker Desktop)
* [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) CLI tool (optional, for local standalone migrations outside docker)

## Architecture
```
.
├── Dockerfile             # Multi-stage build for Go API
├── docker-compose.yml     # Local/VPS environment containing DB and API
├── cmd/server/main.go     # Application Entrypoint
├── internal/
│   ├── auth/              # Authentication & Authorization logic
│   ├── config/            # Environment variable loading
│   ├── db/                # PostgreSQL pgxpool wrapper
│   ├── domain/            # Core models, interfaces, repositories
│   ├── http/              # API Server, Routes, Middlewares
│   └── logic/             # Business Logic (Services)
└── migrations/            # golang-migrate raw SQL scripts
```

## Running the Project Locally (OrbStack/Docker)

1. Make sure OrbStack is running.
2. Ensure you have the `.env` file from the example:
   ```bash
   cp .env.example .env
   ```
3. Boot the environment:
   ```bash
   docker compose up --build
   ```
4. Verify the API is running by hitting the health check:
   ```bash
   curl http://localhost:8081/health
   ```
   *Expected Output: `{"status": "ok", "db": "connected"}`*

## Migrations
The initial HR/Tenancy schema is located in `migrations/`.

If you prefer to run migrations manually against the DB:
```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5433/dml_db?sslmode=disable" up
```
