# OLX API

A Go API project with PostgreSQL connectivity, SQL migrations, and an HTTP health
endpoint. The listings schema is in place; listing endpoints are not implemented
yet.

## Current status

Implemented:

- Configuration loading with `cleanenv`, with file selection and environment
  variable overrides.
- PostgreSQL connection pooling through `database/sql` and the `pgx` driver.
- A database connectivity check before the HTTP server starts.
- Versioned SQL migrations using `golang-migrate` and a separate migration command.
- A `listings` table with UUID identifiers, listing details, and creation timestamps.
- HTTP routing with `http.NewServeMux()` and a JSON `GET /healthz` response.
- Make targets for building, running, and migrating.

`cmd/api/main.go` loads configuration, opens the database connection pool, and
pings PostgreSQL with a five-second timeout. A connection failure terminates
startup. On success, it starts the HTTP server at the configured address and port.
The pool allows up to 25 open and 25 idle connections, with a five-minute maximum
connection lifetime.

The API does not run migrations automatically. There are no listing handlers,
database queries for listings, authentication, or automated tests yet.

## Requirements

- Go 1.27.2 or newer, as declared in `go.mod`.
- A running PostgreSQL database and a connection URL. The database must support
  `gen_random_uuid()`, which the initial migration uses for listing IDs. The
  migration does not install extensions.
- A database role with permission to create the schema objects and, when rolling
  back, drop them.
- GNU Make for the Makefile commands below. You can also use Go directly.

## Getting started

Create a PostgreSQL database and role first. The example below uses database
`olx_api`, user `olx`, and password `olx`; replace these with your local connection
details. Migrations create tables inside an existing database, not the database
itself.

From the repository root, create `config.yaml` using the example in
[Configuration](#configuration), then download dependencies, apply migrations,
and start the API:

```sh
go mod download
CONFIG_PATH=./config.yaml make migrate-up
CONFIG_PATH=./config.yaml make sprint
```

`make sprint` builds `bin/api` and starts the server. Startup logs the configured
listen address after the database connection succeeds. Press `Ctrl+C` to stop the
server. If migrations are already applied, see [Database migrations](#database-migrations)
for the current handling of the `no change` result.

To build and run separately:

```sh
make build
CONFIG_PATH=./config.yaml make run
```

`make run` runs the existing binary; run `make build` again after changing code.
Without Make, use:

```sh
go run ./cmd/api -config_path=./config.yaml
```

With the example configuration, check the server from another terminal:

```sh
curl -i http://127.0.0.1:8080/healthz
```

The expected response is `200 OK` with `Content-Type: application/json` and this
body:

```json
{"status":"ok"}
```

The handler reports that the HTTP server is responsive; it does not query the
database. PostgreSQL is checked only during startup, so this endpoint can remain
healthy if database connectivity is later lost. Unregistered paths return
`404 Not Found`, and unsupported methods on `/healthz` return
`405 Method Not Allowed`. Go's `GET` route also accepts `HEAD /healthz`.

## Configuration

The configuration loader is `MustLoad()` in `internal/config/config.go`. It returns
an `*AppConfig` and terminates the process with a logged error if the configuration
path is missing, the file cannot be read, or required settings are missing.

Create a local `config.yaml` in the repository root with the following contents:

```yaml
env: local
db_url: "postgres://olx:olx@127.0.0.1:5432/olx_api?sslmode=disable"
http_server:
  address: "127.0.0.1"
  port: "8080"
```

`config.yaml` is ignored by Git. These values are examples; the loader does not
provide defaults. The connection URL shown is for local development; use the
connection details and SSL settings required by your database for deployment.

### Configuration file selection

When `MustLoad()` is called, it selects the configuration file in this order:

1. A nonempty `CONFIG_PATH` environment variable.
2. The `-config_path` command-line flag, if `CONFIG_PATH` is empty or unset.

There is no default file path. Relative paths are resolved from the process's
working directory.

Either of these commands selects the local configuration file and starts the
server:

```sh
CONFIG_PATH=./config.yaml go run ./cmd/api
go run ./cmd/api -config_path=./config.yaml
```

### Settings

The loader reads the file first, then applies environment variable overrides.
All values below are strings.

| YAML key | Environment variable | Validation | Purpose |
| --- | --- | --- | --- |
| `env` | `ENV` | Required | Application environment label. |
| `db_url` | `DB_URL` | Not enforced by the loader | PostgreSQL connection URL used by both commands. |
| `http_server.address` | `HTTP_ADDRESS` | Required | HTTP server listen address. |
| `http_server.port` | `HTTP_PORT` | Required before the `PORT` override | HTTP server listen port. |

The required fields can be supplied in the file or through their environment
variables, but a readable configuration file is still required. The nested server
settings use the `HTTP_` environment variable prefix. `env` is currently only a
configuration label; it does not change application behavior.

The `DbUrl` field currently has a misspelled `env-requried` tag, so the loader does
not enforce a database URL. Set `db_url` or `DB_URL` explicitly for both the API
and migration command.

A nonempty `PORT` environment variable overrides the port after configuration
loading, taking precedence over both `HTTP_PORT` and `http_server.port`. Because
required-field validation happens first, `PORT` alone cannot replace a missing
`http_server.port`/`HTTP_PORT` setting.

For example, to override the port:

```sh
CONFIG_PATH=./config.yaml HTTP_PORT=9090 go run ./cmd/api
```

For a deployment that supplies `PORT`, use `HTTP_ADDRESS=0.0.0.0` to listen on all
IPv4 interfaces, and retain a port value in the file or `HTTP_PORT` to satisfy
configuration validation.

## Database migrations

Run migration commands from the repository root: the migration source is the
relative path `file://migrations`. Both commands use the same application
configuration, including the required environment and HTTP settings.

```sh
CONFIG_PATH=./config.yaml make migrate-up
```

This builds `bin/migrate` and applies all pending up migrations. The initial
migration creates `listings` with these columns:

| Column | PostgreSQL type | Constraints / default |
| --- | --- | --- |
| `id` | `UUID` | Primary key; defaults to `gen_random_uuid()`. |
| `title` | `TEXT` | Not null. |
| `description` | `TEXT` | Not null. |
| `price` | `BIGINT` | Not null; no currency or unit is defined yet. |
| `city` | `TEXT` | Not null. |
| `created_at` | `TIMESTAMPTZ` | Not null; defaults to `NOW()`. |

To roll back **all applied migrations**, use:

```sh
CONFIG_PATH=./config.yaml make migrate-down
```

The current down migration drops `listings`, deleting its data. The command does
not prompt for confirmation or support rolling back a single step.

Without Make:

```sh
CONFIG_PATH=./config.yaml go run ./cmd/migrate up
CONFIG_PATH=./config.yaml go run ./cmd/migrate down
```

Use `CONFIG_PATH` for migrations. The current migration command reads the action
from `os.Args[1]`, while the shared configuration loader uses Go's flag parser;
`-config_path` does not work with the positional `up`/`down` action.

The runner currently treats every migration error as fatal, including
`migrate.ErrNoChange`. Repeating `up` when the schema is current, or `down` when
there is nothing to roll back, logs `no change` and exits with a nonzero status.

## Make targets

| Target | Action |
| --- | --- |
| `make build` | Build the API binary at `bin/api`. |
| `make run` | Run the existing `bin/api` binary. |
| `make sprint` | Build and run the API. |
| `make migrate-up` | Build `bin/migrate` and apply all pending migrations. |
| `make migrate-down` | Build `bin/migrate` and roll back all applied migrations. |

## Project structure

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go          # Configuration, database check, and HTTP startup
│   └── migrate/
│       └── main.go          # Migration runner (up/down)
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration types and MustLoad
│   ├── db/
│   │   └── db.go            # PostgreSQL pool setup and startup ping
│   └── handlers/
│       └── healthz.go       # JSON health response
├── migrations/
│   ├── 000001_create_listings_table.up.sql
│   └── 000001_create_listings_table.down.sql
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

The generated `bin/` directory is ignored by Git. The local `config.yaml` is also
ignored and must be created separately.

## Development checks

```sh
go test ./...
make build
go build -o bin/migrate ./cmd/migrate
```

There are no test files yet, so `go test ./...` currently checks that the packages
compile. These checks do not require a running database; applying migrations and
starting the API do.
