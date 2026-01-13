# ==================================================================================== #
## -------
## HELPERS
## -------
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## confirm: prompt before running a destructive action
.PHONY: confirm
confirm:
	@printf "Are you sure? [y/N] " && read ans && [ "$${ans:-N}" = y ]

# ==================================================================================== #
## ---------------------
## ENVIRONMENT VARIABLES
## ---------------------
# ==================================================================================== #

ENV_FILE := .env
ENV_LOAD := set -a; . "$(ENV_FILE)"; set +a

# Optionally, provide a .env.test for manual test-driving the application

main_package_path   ?= ./cmd/dugout
binary_name         ?= dugout

## env/check: fail if .env is missing
.PHONY: env/check
env/check:
	@test -f "$(ENV_FILE)" || ( \
		echo "Refusing: $(ENV_FILE) not found. Create it (or copy from .env.example)." >&2; \
		exit 1; \
	)

# ==================================================================================== #
## ---------------
## QUALITY CONTROL
## ---------------
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: fmt-check mod-tidy-check mod-verify vet staticcheck test/race vulncheck
	@echo "Audit complete."

## mod-tidy-check: fail if go.mod/go.sum are not tidy
.PHONY: mod-tidy-check
mod-tidy-check:
	@echo "Running tidy check ..."
	@go mod tidy -diff
	@echo "... complete.\n"	

## mod-verify: fail if module dependencies cannot be verified
.PHONY: mod-verify
mod-verify:
	@echo "Running mod verify ..."
	@go mod verify
	@echo "... complete.\n"
	
## fmt-check: fail if gofmt would make changes (reports files)
.PHONY: fmt-check
fmt-check:
	@echo "Running gofmt check ..."
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "Refusing: gofmt required on:" >&2; \
		echo "$$files" >&2; \
		exit 1; \
	fi
	@echo "... complete.\n"

## vet: run go vet
.PHONY: vet
vet:
	@echo "Running go vet ..."
	@go vet ./...
	@echo "... complete.\n"	
	
## staticcheck: run staticcheck
.PHONY: staticcheck
staticcheck:
	@echo "Running staticcheck ..."
	@go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	@echo "... complete.\n"	
	
## vulncheck: run govulncheck
.PHONY: vulncheck
vulncheck:
	@echo "Running vulncheck ..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@echo "... complete.\n"	
	
## test: run tests
.PHONY: test
test:
	@echo "Running tests ..."
	@go test -buildvcs ./...
	@echo "... complete.\n"
	
## test/race: run tests with race detector
.PHONY: test/race
test/race:
	@echo "Running tests with race detector ..."
	@go test -race -buildvcs ./...
	@echo "... complete.\n"	
	
## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	@echo "Running tests and displaying coverage ..."
	@go test -buildvcs -coverprofile=/tmp/coverage.out ./...
	@go tool cover -html=/tmp/coverage.out
	@echo "... complete.\n"
	
## test/integration: run integration tests against migrated test DB
.PHONY: test/integration
test/integration: env/check db/test/migrate/up
	@echo "Running integration tests ..."
	@$(ENV_LOAD); \
	go test -buildvcs -tags=integration ./...
	@echo "... complete.\n"
	
## upgradeable: list direct dependencies that have upgrades available
.PHONY: upgradeable
upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest

# ==================================================================================== #
## ----------------------
## CONTINUOUS INTEGRATION
## ----------------------
# ==================================================================================== #

# Pinned tool versions for CI reproducibility (Make vars => lowercase).
staticcheck_version  ?= v0.6.0
govulncheck_version  ?= v1.1.4

## ci: reproducible, pinned quality gate for GitHub Actions
.PHONY: ci
ci: fmt-check mod-tidy-check mod-verify vet ci/staticcheck ci/vulncheck test
	@echo "CI check complete."

## ci/staticcheck: run staticcheck (pinned)
.PHONY: ci/staticcheck
ci/staticcheck:
	@echo "Running staticcheck ..."
	@go run honnef.co/go/tools/cmd/staticcheck@$(staticcheck_version) -checks=all,-ST1000,-U1000 ./...
	@echo "... complete.\n"

## ci/vulncheck: run govulncheck (pinned)
.PHONY: ci/vulncheck
ci/vulncheck:
	@echo "Running vulncheck ..."
	@go run golang.org/x/vuln/cmd/govulncheck@$(govulncheck_version) ./...
	@echo "... complete.\n"

# ==================================================================================== #
## -----------
## DEVELOPMENT
## -----------
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
	
## run/live: run the application with reloading on file changes (optional ARGS passthrough)
.PHONY: run/live
run/live: env/check
	@echo "Running $(binary_name) with automatic refresh on file changes..."
	@$(ENV_LOAD); \
	go run github.com/cosmtrek/air@latest \
		--build.cmd "$(MAKE) --no-print-directory build" \
		--build.bin "/tmp/bin/$(binary_name)" \
		--build.args_bin "$(ARGS)" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"

# Suggested additional targets:
# - run/debug: run with debug-oriented flags (override via ARGS)
# - run/test: create .env.test with substitute variables

# ==================================================================================== #
## --------------
## GIT AND GITHUB
## --------------
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

## require-upstream: fail unless current branch has an upstream tracking branch
.PHONY: require-upstream
require-upstream:
	@up="$$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true)"; \
	if [ -z "$$up" ]; then \
		echo "Refusing: no upstream set for this branch. Run 'make push/u' first." >&2; \
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

## up-to-date: fail unless local HEAD matches origin/main
.PHONY: up-to-date
up-to-date: on-main
	@git fetch origin; \
	if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse origin/main)" ]; then \
		echo "Refusing: local HEAD does not match origin/main." >&2; \
		echo "Hint: run 'make sync/main'." >&2; \
		exit 1; \
	fi

## repair/main: reset local main to origin/main (keeps a backup branch)
.PHONY: repair/main
repair/main: confirm require-clean on-main
	@set -e; \
	git fetch origin; \
	backup="backup/main-local-$$(date +%Y%m%d-%H%M%S)"; \
	echo "Saving current main to '$$backup'..."; \
	git branch "$$backup" HEAD; \
	echo "Resetting main to origin/main..."; \
	git reset --hard origin/main

## sync/main: fast-forward main from origin/main (no confirm; safe to call from other targets)
.PHONY: sync/main
sync/main: require-clean
	@echo "Syncing main from origin/main ..."
	@git switch main >/dev/null
	@git pull --ff-only

## sync/branch: rebase onto upstream then origin/main, audit, and publish (force-with-lease)
.PHONY: sync/branch
sync/branch: confirm require-clean on-feature require-upstream
	@echo "Syncing branch ..."
	@git fetch origin
	@git rebase @{u}
	@git rebase origin/main
	@$(MAKE) --no-print-directory audit
	@git push --force-with-lease

## sync: convenience alias for sync/branch
.PHONY: sync
sync: sync/branch

## rebase/upstream: rebase current branch onto its upstream (keeps branches linear)
.PHONY: rebase/upstream
rebase/upstream: confirm require-clean on-feature require-upstream
	@echo "Rebasing branch onto upstream ..."
	@git fetch origin
	@git rebase @{u}

## rebase/main: rebase current branch onto origin/main (keeps branches linear)
.PHONY: rebase/main
rebase/main: confirm require-clean on-feature
	@echo "Rebasing main onto origin/main ..."
	@git fetch origin
	@git rebase origin/main

## branch/new: create and switch to a new work branch (from freshly synced main)
.PHONY: branch/new
branch/new: confirm require-clean
	@set -e; \
	$(MAKE) --no-print-directory on-main; \
	$(MAKE) --no-print-directory sync/main; \
	printf "Branch type (feature|fix|refactor|chore|docs): " ; \
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
	git switch -c "$$branch"

## branch/cleanup: sync main (from GitHub) and delete the current branch locally (squash-merge safe)
.PHONY: branch/cleanup
branch/cleanup: confirm require-clean
	@set -e; \
	branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" = "main" ]; then \
		echo "error: refusing to delete 'main'"; \
		exit 1; \
	fi; \
	echo "Cleaning up branch '$$branch'..."; \
	$(MAKE) --no-print-directory sync/main; \
	git branch -D "$$branch"

## push: fast-forward-only push (cheap pushes); refuse if it would be non-fast-forward
.PHONY: push
push: on-feature require-upstream require-clean
	@git push

## push/u: push current branch and set upstream to origin
.PHONY: push/u
push/u: on-feature require-clean
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	git push -u origin "$$branch"
	
## pr/create: create a GitHub PR for the current branch
.PHONY: pr/create
pr/create: confirm audit on-feature
	@gh pr create --fill --editor

## pr/view: open the current PR in the browser
.PHONY: pr/view
pr/view: on-feature
	@gh pr view --web

# ==================================================================================== #
## --------
## DATABASE
## --------
# ==================================================================================== #

bootstrap_sql  ?= ./database/bootstrap.sql
migrations_dir ?= ./database/migrations

# DSN templates (expanded by the shell after $(ENV_LOAD))
dsn_admin         = host=$$DB_HOST port=$$DB_PORT dbname=postgres user=$$DB_USER_ADMIN sslmode=$$DB_SSLMODE
dsn_app_dev       = host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME user=$$DB_USER_APP sslmode=$$DB_SSLMODE
dsn_migrator_dev  = host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME user=$$DB_USER_MIGRATOR sslmode=$$DB_SSLMODE
dsn_app_test      = host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME_TEST user=$$DB_USER_APP sslmode=$$DB_SSLMODE
dsn_migrator_test = host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME_TEST user=$$DB_USER_MIGRATOR sslmode=$$DB_SSLMODE

## db/tools/check: verify required tooling is installed
.PHONY: db/tools/check
db/tools/check:
	@command -v sqlc >/dev/null 2>&1 || { echo "Refusing: sqlc not found. Install it and try again." >&2; exit 1; }
	@command -v goose >/dev/null 2>&1 || { echo "Refusing: goose not found. Install it and try again." >&2; exit 1; }
	@command -v psql >/dev/null 2>&1 || { echo "Refusing: psql not found. Install Postgres client tools and try again." >&2; exit 1; }

## db/check: fail if required DB env vars are not set
.PHONY: db/check
db/check: env/check db/tools/check
	@$(ENV_LOAD); \
	test -n "$$DB_HOST" || (echo "Refusing: DB_HOST is not set." >&2; exit 1); \
	test -n "$$DB_PORT" || (echo "Refusing: DB_PORT is not set." >&2; exit 1); \
	test -n "$$DB_SSLMODE" || (echo "Refusing: DB_SSLMODE is not set." >&2; exit 1); \
	test -n "$$DB_NAME" || (echo "Refusing: DB_NAME is not set." >&2; exit 1); \
	test -n "$$DB_NAME_TEST" || (echo "Refusing: DB_NAME_TEST is not set." >&2; exit 1); \
	test -n "$$DB_USER_ADMIN" || (echo "Refusing: DB_USER_ADMIN is not set." >&2; exit 1); \
	test -n "$$DB_USER_MIGRATOR" || (echo "Refusing: DB_USER_MIGRATOR is not set." >&2; exit 1); \
	test -n "$$DB_USER_APP" || (echo "Refusing: DB_USER_APP is not set." >&2; exit 1)

## db/bootstrap: create roles and databases (safe-ish to re-run)
.PHONY: db/bootstrap
db/bootstrap: db/check
	@test -f "$(bootstrap_sql)" || (echo "Refusing: $(bootstrap_sql) not found." >&2; exit 1)
	@$(ENV_LOAD); \
	out="$$(mktemp)"; \
	if psql "$(dsn_admin)" -v ON_ERROR_STOP=1 -f "$(bootstrap_sql)" >"$$out" 2>&1; then \
		cat "$$out"; rm -f "$$out"; exit 0; \
	fi; \
	allowed_re="database \"($$DB_NAME|$$DB_NAME_TEST)\" already exists"; \
	other_err="$$(grep -E '^psql:.*ERROR:' "$$out" | grep -Ev "$$allowed_re" || true)"; \
	if grep -Eq 'ERROR:.*database ".*" already exists' "$$out" && [ -z "$$other_err" ]; then \
		cat "$$out"; rm -f "$$out"; exit 0; \
	fi; \
	cat "$$out" >&2; rm -f "$$out"; exit 1

## db/sqlc: generate Go code from SQL queries
.PHONY: db/sqlc
db/sqlc: db/tools/check
	@sqlc generate
	
## db/gen: run migrations then regenerate sqlc (common dev workflow)
.PHONY: db/gen
db/gen: db/migrate/up db/sqlc

## db/connect: connect to the dev database with psql
.PHONY: db/connect
db/connect: db/check
	@$(ENV_LOAD); \
	psql "$(dsn_app_dev)"

## db/ping: verify database connectivity
.PHONY: db/ping
db/ping: db/check
	@$(ENV_LOAD); \
	psql "$(dsn_app_dev)" -c 'select 1' >/dev/null

## db/migrate/new name=...: create a new migration file (sequential numbering)
.PHONY: db/migrate/new
db/migrate/new: db/tools/check
	@test -n "$(name)" || (echo "Usage: make db/migrate/new name=<migration_name>" >&2; exit 1)
	@test -d "$(migrations_dir)" || (echo "Refusing: migrations_dir not found: $(migrations_dir)" >&2; exit 1)
	@goose -s -dir "$(migrations_dir)" create "$(name)" sql

## db/migrate/status: show migration status
.PHONY: db/migrate/status
db/migrate/status: db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_dev)" status

## db/migrate/up: apply all up migrations
.PHONY: db/migrate/up
db/migrate/up: db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_dev)" up

## db/migrate/down: roll back the most recent migration
.PHONY: db/migrate/down
db/migrate/down: confirm db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_dev)" down

## db/migrate/reset: rollback all migrations, then migrate up (DESTRUCTIVE)
.PHONY: db/migrate/reset
db/migrate/reset: confirm db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_dev)" reset

## db/migrate/version: print current migration version
.PHONY: db/migrate/version
db/migrate/version: db/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_dev)" version

# ==================================================================================== #
## -------------
## TEST DATABASE
## -------------
# ==================================================================================== #

## db/test/check: fail if test DB is not set (and guard against pointing at dev DB)
.PHONY: db/test/check
db/test/check: db/check
	@$(ENV_LOAD); \
	test -n "$$DB_NAME_TEST" || (echo "Refusing: DB_NAME_TEST is not set." >&2; exit 1); \
	test "$$DB_NAME_TEST" != "$$DB_NAME" || (echo "Refusing: DB_NAME_TEST must not equal DB_NAME." >&2; exit 1)

## db/test/connect: connect to the test database with psql
.PHONY: db/test/connect
db/test/connect: db/test/check
	@$(ENV_LOAD); \
	psql "$(dsn_app_test)"

## db/test/ping: verify test database connectivity
.PHONY: db/test/ping
db/test/ping: db/test/check
	@$(ENV_LOAD); \
	psql "$(dsn_app_test)" -c 'select 1' >/dev/null

## db/test/migrate/status: show migration status (test DB)
.PHONY: db/test/migrate/status
db/test/migrate/status: db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_test)" status

## db/test/migrate/up: apply all up migrations (test DB)
.PHONY: db/test/migrate/up
db/test/migrate/up: db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_test)" up

## db/test/migrate/down: roll back the most recent migration (test DB)
.PHONY: db/test/migrate/down
db/test/migrate/down: confirm db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_test)" down

## db/test/migrate/reset: rollback all migrations, then migrate up (DESTRUCTIVE, test DB)
.PHONY: db/test/migrate/reset
db/test/migrate/reset: confirm db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_test)" reset

## db/test/migrate/version: print current migration version (test DB)
.PHONY: db/test/migrate/version
db/test/migrate/version: db/test/check
	@$(ENV_LOAD); \
	goose -dir "$(migrations_dir)" postgres "$(dsn_migrator_test)" version

# ==================================================================================== #
## ----------
## DEPLOYMENT
## ----------
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
production/connect: env/check
	@$(ENV_LOAD); \
	test -n "$$PRODUCTION_HOST_IP" || (echo "Refusing: PRODUCTION_HOST_IP is not set." >&2; exit 1); \
	test -n "$$PRODUCTION_SSH_USER" || (echo "Refusing: PRODUCTION_SSH_USER is not set." >&2; exit 1); \
	ssh "$$PRODUCTION_SSH_USER@$$PRODUCTION_HOST_IP"

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

