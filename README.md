# Dugout
*A Go-based toolkit for managing baseball rosters and stats.*

**Dugout** is a monorepo for a collection of baseball organization management tools and APIs written in Go. It begins with a backend service for roster management and player statistics, forming the foundation for future CLI and web interfaces.

---

## Overview

Dugout will eventually include:
- **Roster Manager API** -- A REST service for managing organizations, teams, and players.
- **CLI Tools** -- Local utilities for administrative and data import tasks.
- **Shared Packages** -- Common domain models, database code, and testing utilities.

---

## Tech Stack

- Language: Go (Golang)
- Database: PostgreSQL
- Migrations: Goose
- Codegen: sqlc
- Environment: `.env` for configuration
- Testing: Go test framework

---

## Environment Setup

Before running the application, copy the example file and update it with your local settings:

```bash
cp .env.example .env
```

Your `.env` file should include the following variables:

```dotenv
POSTGRES_USER=<user>
POSTGRES_PASSWORD=<password>
POSTGRES_DB=<database>

DB_URL=postgres://<user>:<password>@localhost:5432/<database>?sslmode=disable
```

These values are used by the database container, migrations scripts, and application startup.

## Development Setup

```bash
# clone the repo
git clone https://github.com/spcameron/dugout.git
cd dugout

# initialize the database and migrations
make db
make migrate

# run tests
make test
```

---

## Contributing

We welcome your contributions! To get started:
- Fork the repository on GitHub
- Create a feature branch:
```bash
git switch -c feature/<short-description>
```
- Commit your changes with clear messages
- Push your branch and open a pull request

Please keep PRs focused and concise. Before submitting, run `make test` and ensure all tests pass.

## License

This project is licensed under the **MIT License.**


