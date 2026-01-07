# Dugout
*A Go-based toolkit for managing baseball rosters and stats*

**Dugout** is a monorepo for a collection of baseball organization management tools and APIs written in Go. It provides a backend service for roster management and player statistics, forming the foundation for future CLI and web interfaces.

---

## Overview

Dugout will eventually include:
- **Roster Manager API** -- A REST service for managing organizations, teams, and players.
- **CLI Tools** -- Local utilities for administrative and data import tasks.
- **Web Application** -- A user-facing browser interface for interacting with the API.
- **Shared Packages** -- Common domain models, database code, and testing utilities.

---

## Tech Stack

- Language: Go (Golang)
- Database: PostgreSQL
- Migrations: Goose
- Codegen: sqlc
- Routing: chi
- Environment: `.env` for configuration
- Testing: Go test framework
- Web UI:
    - htmx
    - templ
    - Tailwind CSS
    - Alpine.js

---

## Development Setup

### Prerequisites

You will need the following installed locally:
- Go (version 1.25 or higher)
- PostgreSQL (version 16)
- `psql` client
- `make`

Please also make sure `goose` and `sqlc` are installed.

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Clone the repository

```bash
# clone the repo
git clone https://github.com/spcameron/dugout.git
cd dugout
```

This repository provides a **first-class Makefile.** It is not merely a convenience layer; it is the intended interface for:
- database bootstrapping
- migrations
- code generation
- local execution

Contributors should expect to interact with the project primarily through `make`.

---

## Environment Setup

### 1. Create a `.env` file

Before running any database-related commands, copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` as needed. At minimum, it should define:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_SSLMODE=disable

DB_NAME=dugout_dev
DB_NAME_TEST=dugout_test

DB_USER_ADMIN=your_local_pg_admin_role
DB_USER_MIGRATOR=dugout_migrator
DB_USER_APP=dugout_app
```

These values define **connection facts**, not credentials.

`.env` intentionally does **not** contain passwords.

### 2. Bootstrap the database

Initialize the required PostreSQL roles and databases:

```bash
make db/bootstrap
```

This step creates the application and migrator roles, creates development and test databases, and assigns ownership and base permissions. It is safe to re-run.

**Note:** On many macOS/Homebrew installations, the default Postgres superuser role is your OS username, not `postgres`. You can list local roles with `psql -d postgres -c '\du'`. Set `DB_USER_ADMIN` in `.env` accordingly.

### 3. Configure database authentication (recommended)

This project relies on PostgreSQL role separation (admin, migrator, app) and expects authentication to be configured locally, outside the repository.

After bootstrapping has created the roles, set local passwords for the non-admin roles:

```bash
app_pw="$(openssl rand -base64 24)"
migrator_pw="$(openssl rand -base64 24)"

psql -d postgres -c "ALTER ROLE dugout_app WITH PASSWORD '$app_pw';"
psql -d postgres -c "ALTER ROLE dugout_migrator WITH PASSWORD '$migrator_pw';"
```

Choose any secure local method you prefer for generating passwords. These values are never committed to the repository.

To avoid password prompts during development, use PostgreSQL’s machine-local password file, `~/.pgpass`.

Create or edit the file at:

```bash
~/.pgpass
```

Add entries for your local development and test databases:

```bash
localhost:5432:dugout_dev:dugout_app:<app_password>
localhost:5432:dugout_dev:dugout_migrator:<migrator_password>
localhost:5432:dugout_test:dugout_app:<app_password>
localhost:5432:dugout_test:dugout_migrator:<migrator_password>
```

Then restrict permissions:

```bash
chmod 600 ~/.pgpass
```

Alternatives:

Advanced users may prefer peer authentication or environment variables such as PGPASSWORD.

The Makefile supports these, but .pgpass provides the smoothest local experience.

### 4. Run migrations

Apply schema migrations to both databases:

```bash
make db/migrate/up
make db/test/migrate/up
```
Migrations are executed using the **migrator role**, not the application role.

### 5. Running the application

After bootstrapping and migrating the databases, you can build and run the application locally using the Makefile.

```bash
make build
```

This target build the application binary to `/tmp/bin/dugout`, loads environment variables from `.env`, and then runs the binary with optional argument passthrough via `ARGS`. For example:

```bash
make run ARGS="--help"
```

You may also wish to run the application with live reload
```bash
make run/live
```

This uses `air` to rebuild and restart the application automatically when files change (Go, templates, SQL, and common web assets). Note: `run/live` installs and runs `air` via `go run ...@latest`. If you prefer a pinned version, install `air` separately and update the Makefile accordingly.

### Common Issues

- If the application fails to start due to missing configuration, confirm that .env exists and matches .env.example.
- If you see database authentication errors, verify:
    - `make db/bootstrap` has been run
    - role passwords are set locally
    - `~/.pgpass` exists and has correct permissions
    - `make db/migrate/up` completes successfully

---

## Contributing

We welcome your contributions! To get started:
- Fork the repository on Github
- Follow the development and environment setup instructions above
- Create a feature branch by running `make branch/new`
- Commit your changes with clear messages
- Push your branch and open a pull request

Please keep PRs focused and concise. Before submitting, run `make audit` and ensure all tests pass.

---

## License

This project is licensed under the **MIT License.**

![CI tests badge](https://github.com/spcameron/dugout/actions/workflows/ci.yml/badge.svg)
