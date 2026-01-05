# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## confirm: prompt before running a destructive action
.PHONY: confirm
confirm:
	@printf 'Are you sure? [y/N] ' && read ans && [ "$${ans:-N}" = y ]

# ==================================================================================== #
# ENVIRONMENT VARIABLES
# ==================================================================================== #

ENV_FILE := .env
ENV_LOAD := set -a; . "$(ENV_FILE)"; set +a

# Optionally, provide a .env.test for manual test-driving the application

# MUST UPDATE THESE PLACEHOLDERS
main_package_path   ?= ./cmd/example
binary_name         ?= example
production_host_ip  ?= xxx.xxx.xx.xxx
production_ssh_user ?= example_user
migrations_dir      ?= ./migrations

## env/check: fail if .env is missing
.PHONY: env/check
env/check:
	@test -f "$(ENV_FILE)" || ( \
		echo "Refusing: $(ENV_FILE) not found. Create it (or copy from .env.example)." >&2; \
		exit 1; \
	)

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: fmt-check mod-tidy-check mod-verify vet staticcheck test/race vulncheck

## mod-tidy-check: fail if go.mod/go.sum are not tidy
.PHONY: mod-tidy-check
mod-tidy-check:
	@go mod tidy -diff
	
## mod-verify: fail if module dependencies cannot be verified
.PHONY: mod-verify
mod-verify:
	@go mod verify
	
## fmt-check: fail if gofmt would make changes (reports files)
.PHONY: fmt-check
fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "Refusing: gofmt required on:" >&2; \
		echo "$$files" >&2; \
		exit 1; \
	fi

## vet: run go vet
.PHONY: vet
vet:
	@go vet ./...
	
## staticcheck: run staticcheck
.PHONY: staticcheck
staticcheck:
	@go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	
## vulncheck: run govulncheck
.PHONY: vulncheck
vulncheck:
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	
## test: run tests
.PHONY: test
test:
	@go test -buildvcs ./...
	
## test/race: run tests with race detector
.PHONY: test/race
test/race:
	@go test -race -buildvcs ./...
	
## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	@go test -buildvcs -coverprofile=/tmp/coverage.out ./...
	@go tool cover -html=/tmp/coverage.out
	
## test/integration: run integration tests against migrated test DB
.PHONY: test/integration
test/integration: env/check db/test/migrate/up
	@$(ENV_LOAD); \
	go test -buildvcs -tags=integration ./...
	
## upgradeable: list direct dependencies that have upgrades available
.PHONY: upgradeable
upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest

# ==================================================================================== #
# CI
# ==================================================================================== #

# Pinned tool versions for CI reproducibility (Make vars => lowercase).
staticcheck_version  ?= v0.6.0
govulncheck_version  ?= v1.1.4

## ci: reproducible, pinned quality gate for GitHub Actions
.PHONY: ci
ci: fmt-check mod-tidy-check mod-verify vet ci/staticcheck ci/vulncheck test

## ci/staticcheck: run staticcheck (pinned)
.PHONY: ci/staticcheck
ci/staticcheck:
	@go run honnef.co/go/tools/cmd/staticcheck@$(staticcheck_version) -checks=all,-ST1000,-U1000 ./...

## ci/vulncheck: run govulncheck (pinned)
.PHONY: ci/vulncheck
ci/vulncheck:
	@go run golang.org/x/vuln/cmd/govulncheck@$(govulncheck_version) ./...

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	@go mod tidy -v
	@go fmt ./...
	
## build: build the application (local)
.PHONY: build
build:
	@mkdir -p /tmp/bin
	@go build -o=/tmp/bin/$(binary_name) $(main_package_path)
	
## build/linux_amd64: build the production binary
.PHONY: build/linux_amd64
build/linux_amd64:
	@mkdir -p /tmp/bin/linux_amd64
	@GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/$(binary_name) $(main_package_path)
	
## run: run the application (optional ARGS passthrough)
.PHONY: run
run: env/check build
	@echo "Running $(binary_name)..."
	@$(ENV_LOAD); \
	/tmp/bin/$(binary_name) $(ARGS)
	
## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live: env/check 
	@echo "Running $(binary_name) with automatic refresh on file changes..."
	@$(ENV_LOAD); \
	go run github.com/cosmtrek/air@latest \
		--build.cmd "$(MAKE) build" --build.bin "/tmp/bin/$(binary_name)" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"
		
# Suggested additional targets:
# - run/debug: run with debug-oriented flags (override via ARGS)
# - run/test: create .env.test with substitute variables

# ==================================================================================== #
# GIT AND GITHUB
# ==================================================================================== #

## require-clean: fail if the Git working tree has uncommitted changes
.PHONY: require-clean
require-clean:
	@status="$$(git status --porcelain)"; \
	if [ -n "$$status" ]; then \
		echo "Refusing: working tree is dirty (commit/stash first)." >&2; \
		echo "$$status" >&2; \
		exit 1; \
	fi

## no-dirty: alias for require-clean (deprecated)
.PHONY: no-dirty
no-dirty: require-clean

## require-upstream: fail unless current branch has an upstream tracking branch
.PHONY: require-upstream
require-upstream:
	@up="$$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true)"; \
	if [ -z "$$up" ]; then \
		echo "Refusing: no upstream set for this branch. Run 'make push/u' first." >&2; \
		exit 1; \
	fi

## up-to-date: fail unless local HEAD matches origin/main
.PHONY: up-to-date
up-to-date: on-main
	@git fetch origin; \
	if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse origin/main)" ]; then \
		echo "Refusing: local HEAD does not match origin/main." >&2; \
		echo "Hint: run 'git pull --ff-only' on main." >&2; \
		exit 1; \
	fi
	
## on-main: fail unless currently on main
.PHONY: on-main
on-main:
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" != "main" ]; then \
		echo "Refusing: not on main (current: $$branch)." >&2; \
		exit 1; \
	fi

## on-feature: fail unless currently on a non-main branch with an approved prefix
.PHONY: on-feature
on-feature:
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" = "HEAD" ]; then \
		echo "Refusing: detached HEAD (checkout a branch)." >&2; \
		exit 1; \
	fi; \
	case "$$branch" in \
		feature/*|fix/*|refactor/*|chore/*|docs/*) ;; \
		main) echo "Refusing: on main; use a work branch (feature/*, fix/*, refactor/*, chore/*, docs/*)." >&2; exit 1 ;; \
		*) echo "Refusing: branch '$$branch' not in allowed prefixes (feature/, fix/, refactor/, chore/, docs/)." >&2; exit 1 ;; \
	esac

## sync/main: fast-forward main from origin/main
.PHONY: sync/main
sync/main: confirm require-clean
	@git switch main
	@git pull --ff-only
	
## branch/new: create and switch to a new work branch
.PHONY: branch/new
branch/new: confirm require-clean on-main
	@printf "Branch type (feature|fix|refactor|chore|docs): " ; \
	read type ; \
	case "$$type" in feature|fix|refactor|chore|docs) ;; \
		*) echo "Refusing: invalid type '$$type'." >&2; exit 1 ;; \
	esac ; \
	printf "Slug (lowercase, digits, hyphens; e.g. add-login): " ; \
	read slug ; \
	case "$$slug" in ""|[^a-z0-9]*|*[!a-z0-9-]*) \
		echo "Refusing: invalid slug '$$slug'." >&2; exit 1 ;; \
	esac ; \
	branch="$$type/$$slug" ; \
	echo "Creating branch $$branch from main..." ; \
	git pull --ff-only ; \
	git switch -c "$$branch"
	
## rebase/main: rebase current branch onto origin/main
.PHONY: rebase/main
rebase/main: confirm require-clean on-feature
	@git fetch origin
	@git rebase origin/main
	
## push: push changes to the remote Git repository
.PHONY: push
push: on-feature require-upstream
	@git push
	
## push/u: push current branch and set upstream to origin
.PHONY: push/u
push/u: on-feature
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	git push -u origin "$$branch"
	
## push/pr: rebase and update an open PR branch (force-with-lease)
.PHONY: push/pr
push/pr: confirm on-feature require-upstream require-clean
	@git fetch origin
	@git rebase origin/main
	@$(MAKE) audit
	@git push --force-with-lease
	
## pr/create: create a GitHub PR for the current branch
.PHONY: pr/create
pr/create: confirm audit on-feature
	@gh pr create
	
## pr/view: open the current PR in the browser
.PHONY: pr/view
pr/view: on-feature
	@gh pr view --web

## cleanup/feature: switch to main and delete the current feature branch locally
.PHONY: cleanup/feature
cleanup/feature: confirm require-clean on-feature
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	git switch main; \
	git pull --ff-only; \
	git branch -d "$$branch"

# ==================================================================================== #
# DATABASE (Postgres + goose + sqlc)
# ==================================================================================== #

## tools/check: verify required tooling is installed
.PHONY: tools/check
tools/check:
	@command -v sqlc >/dev/null 2>&1 || { echo "Refusing: sqlc not found. Install it and try again." >&2; exit 1; }
	@command -v goose >/dev/null 2>&1 || { echo "Refusing: goose not found. Install it and try again." >&2; exit 1; }

## sqlc: generate Go code from SQL queries
.PHONY: sqlc
sqlc: tools/check
	@sqlc generate
	
## db/gen: run migrations then regenerate sqlc (common dev workflow)
.PHONY: db/gen
db/gen: db/migrate/up sqlc

## db/check: fail if DB_DSN is not set
.PHONY: db/check
db/check: env/check
	@$(ENV_LOAD); \
	test -n "$(DB_DSN)" || (echo "Refusing: DB_DSN is not set." >&2; exit 1)

## db/connect: connect to the dev database with psql
.PHONY: db/connect
db/connect: db/check
	@$(ENV_LOAD); \
	psql "$(DB_DSN)"

## db/ping: verify database connectivity
.PHONY: db/ping
db/ping: db/check
	@$(ENV_LOAD); \
	psql "$(DB_DSN)" -c 'select 1' >/dev/null

## db/migrate/new name=...: create a new migration file
.PHONY: db/migrate/new
db/migrate/new: tools/check
	@test -n "$(name)" || (echo "Usage: make db/migrate/new name=<migration_name>" >&2; exit 1)
	@goose -dir "$(migrations_dir)" create "$(name)" sql

## db/migrate/status: show migration status
.PHONY: db/migrate/status
db/migrate/status: tools/check db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(DB_DSN)" status

## db/migrate/up: apply all up migrations
.PHONY: db/migrate/up
db/migrate/up: tools/check db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(DB_DSN)" up

## db/migrate/down: roll back the most recent migration
.PHONY: db/migrate/down
db/migrate/down: confirm tools/check db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(DB_DSN)" down

## db/migrate/reset: rollback all migrations, then migrate up (DESTRUCTIVE)
.PHONY: db/migrate/reset
db/migrate/reset: confirm tools/check db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(DB_DSN)" reset

## db/migrate/version: print current migration version
.PHONY: db/migrate/version
db/migrate/version: tools/check db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(DB_DSN)" version

# ==================================================================================== #
# TEST DATABASE (Postgres + goose + sqlc)
# ==================================================================================== #

## db/test/check: fail if TEST_DB_DSN is not set (and guard against pointing at dev DSN)
.PHONY: db/test/check
db/test/check: env/check
	@$(ENV_LOAD); \
	test -n "$(TEST_DB_DSN)" || (echo "Refusing: TEST_DB_DSN is not set." >&2; exit 1); \
	test "$(TEST_DB_DSN)" != "$(DB_DSN)" || (echo "Refusing: TEST_DB_DSN must not equal DB_DSN." >&2; exit 1)

## db/test/connect: connect to the test database with psql
.PHONY: db/test/connect
db/test/connect: db/test/check
	@$(ENV_LOAD); \
	psql "$(TEST_DB_DSN)"

## db/test/ping: verify test database connectivity
.PHONY: db/test/ping
db/test/ping: db/test/check
	@$(ENV_LOAD); \
	psql "$(TEST_DB_DSN)" -c 'select 1' >/dev/null

## db/test/migrate/status: show migration status (test DB)
.PHONY: db/test/migrate/status
db/test/migrate/status: tools/check db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(TEST_DB_DSN)" status

## db/test/migrate/up: apply all up migrations (test DB)
.PHONY: db/test/migrate/up
db/test/migrate/up: tools/check db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(TEST_DB_DSN)" up

## db/test/migrate/down: roll back the most recent migration (test DB)
.PHONY: db/test/migrate/down
db/test/migrate/down: confirm tools/check db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(TEST_DB_DSN)" down

## db/test/migrate/reset: rollback all migrations, then migrate up (DESTRUCTIVE, test DB)
.PHONY: db/test/migrate/reset
db/test/migrate/reset: confirm tools/check db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(TEST_DB_DSN)" reset

## db/test/migrate/version: print current migration version (test DB)
.PHONY: db/test/migrate/version
db/test/migrate/version: tools/check db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(TEST_DB_DSN)" version

# ==================================================================================== #
# DEPLOYMENT
# ==================================================================================== #

## tools/check: verify required tooling is installed
.PHONY: production/tools/check
production/tools/check:
	@command -v upx >/dev/null 2>&1 || { echo "Refusing: upx not found. Install it or remove compression from deploy." >&2; exit 1; }

## predeploy: run checks required before a deployment
.PHONY: predeploy
predeploy: audit require-clean on-main up-to-date confirm
	@echo "Pre-deploy checks passed."

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	@ssh $(production_ssh_user)@"$(production_host_ip)"

## production/deploy: deploy the application to production
.PHONY: production/deploy
production/deploy: predeploy production/tools/check build/linux_amd64
	@upx -5 /tmp/bin/linux_amd64/$(binary_name)
	# Include additional deployment steps here...

# Suggested additional targets:
# - staging/deploy: deploy the application to a staging server
# - production/log: view production logs
# - production/db: connect to production database
# - production/upgrade: update and upgrade software on production server


