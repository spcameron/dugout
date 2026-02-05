# ==================================================================================== #
# COMPLETELY OVER THE TOP ANSI COLOR SUPPORT
# ==================================================================================== #

define log_info
	if [ -t 1 ]; then \
		printf "\033[34m%s\033[0m\n" "$(1)"; \
	else \
		printf "%s\n" "$(1)"; \
	fi
endef

define log_ok
	if [ -t 1 ]; then \
		printf "\033[32m%s\033[0m\n" "$(1)"; \
	else \
		printf "%s\n" "$(1)"; \
	fi
endef

define log_warn
	if [ -t 2 ]; then \
		printf "\033[33m%s\033[0m\n" "$(1)" >&2; \
	else \
		printf "%s\n" "$(1)" >&2; \
	fi
endef

define log_err
	if [ -t 2 ]; then \
		printf "\033[31m%s\033[0m\n" "$(1)" >&2; \
	else \
		printf "%s\n" "$(1)" >&2; \
	fi
endef

define die
	if [ -t 2 ]; then \
		printf "\033[31m%s\033[0m\n" "$(1)" >&2; \
	else \
		printf "%s\n" "$(1)" >&2; \
	fi; \
	exit 1
endef

# ==================================================================================== #
## -------
## HELPERS
## -------
# ==================================================================================== #

## help: print this help message -- OK
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## confirm: prompt before running a destructive action -- OK
.PHONY: confirm
confirm:
	@$(call log_warn,Are you sure? [y/N])
	@printf ">> "
	@read ans && [ "$${ans:-N}" = y ]

# ==================================================================================== #
## ---------------------
## ENVIRONMENT VARIABLES
## ---------------------
# ==================================================================================== #

ENV_FILE := .env
ENV_LOAD := set -e; set -a; . "$(ENV_FILE)"; set +a

# Optionally, provide a .env.test for manual test-driving the application

main_package_path   ?= ./cmd/dugout
binary_name         ?= dugout

## env/check: fail if .env is missing
.PHONY: env/check
env/check:
	@test -f "$(ENV_FILE)" || { $(call die,Refusing: $(ENV_FILE) not found. Create it (or copy from .env.example).); }

# ==================================================================================== #
## ---------------
## QUALITY CONTROL
## ---------------
# ==================================================================================== #

## audit: run quality control checks -- OK
.PHONY: audit
audit: fmt-check mod-tidy-check mod-verify vet staticcheck vulncheck test/race
	@$(call log_ok,Audit complete.)
	@echo

## fmt-check: fail if gofmt would make changes (reports files) -- OK
.PHONY: fmt-check
fmt-check:
	@$(call log_info,Running gofmt check...)
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		$(call log_err,Refusing: gofmt required on:); \
		echo "$$files" >&2; \
		exit 1; \
	fi
	@$(call log_ok,... complete.)
	@echo

## mod-tidy-check: fail if go.mod/go.sum are not tidy -- OK
.PHONY: mod-tidy-check
mod-tidy-check:
	@$(call log_info,Running tidy check...)
	@go mod tidy -diff
	@$(call log_ok,... complete.)
	@echo

## mod-verify: fail if module dependencies cannot be verified -- OK
.PHONY: mod-verify
mod-verify:
	@$(call log_info,Running mod verify...)
	@go mod verify
	@$(call log_ok,... complete.)
	@echo

## vet: run go vet -- OK
.PHONY: vet
vet:
	@$(call log_info,Running go vet...)
	@go vet ./...
	@$(call log_ok,... complete.)
	@echo
	
## staticcheck: run staticcheck -- OK
.PHONY: staticcheck
staticcheck:
	@$(call log_info,Running staticcheck...)
	@go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	@$(call log_ok,... complete.)
	@echo
	
## vulncheck: run govulncheck -- OK
.PHONY: vulncheck
vulncheck:
	@$(call log_info,Running vulncheck...)
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@$(call log_ok,... complete.)
	@echo
	
## test: run tests -- OK
.PHONY: test
test:
	@$(call log_info,Running tests...)
	@go test -buildvcs ./...
	@$(call log_ok,... complete.)
	@echo
	
## test/race: run tests with race detector -- OK
.PHONY: test/race
test/race:
	@$(call log_info,Running tests with race detector...)
	@go test -race -buildvcs ./...
	@$(call log_ok,... complete.)
	@echo
	
## test/cover: run all tests and display coverage -- OK
.PHONY: test/cover
test/cover:
	@$(call log_info,Running tests and displaying coverage...)
	@go test -buildvcs -coverprofile=/tmp/coverage.out ./...
	@go tool cover -html=/tmp/coverage.out
	@$(call log_ok,... complete.)
	@echo
	
## TODO:
## test/integration: run integration tests against migrated test DB
.PHONY: test/integration
test/integration: env/check db/test/migrate/up
	@$(call log_info,Running integration tests...)
	@$(ENV_LOAD); \
	go test -buildvcs -tags=integration ./...
	@$(call log_ok,... complete.)
	@echo
	
## upgradeable: list direct dependencies that have upgrades available -- OK
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

## ci: reproducible, pinned quality gate for GitHub Actions -- OK
.PHONY: ci
ci: fmt-check mod-tidy-check mod-verify vet ci/staticcheck ci/vulncheck test
	@$(call log_ok,CI check complete.)
	@echo

## ci/staticcheck: run staticcheck (pinned) -- OK
.PHONY: ci/staticcheck
ci/staticcheck:
	@$(call log_info,Running staticcheck...)
	@go run honnef.co/go/tools/cmd/staticcheck@$(staticcheck_version) -checks=all,-ST1000,-U1000 ./...
	@$(call log_ok,... complete.)
	@echo

## ci/vulncheck: run govulncheck (pinned) -- OK
.PHONY: ci/vulncheck
ci/vulncheck:
	@$(call log_info,Running vulncheck...)
	@go run golang.org/x/vuln/cmd/govulncheck@$(govulncheck_version) ./...
	@$(call log_ok,... complete.)
	@echo

# ==================================================================================== #
## -----------
## DEVELOPMENT
## -----------
# ==================================================================================== #

## tidy: tidy modfiles and format .go files -- OK
.PHONY: tidy
tidy:
	@$(call log_info,Running tidy...)
	@go mod tidy -v
	@gofmt ./...
	@$(call log_ok,... complete.)
	@echo
	
## build: build the application (local) -- OK
.PHONY: build
build:
	@$(call log_info,Building $(binary_name) \(local\)...)
	@mkdir -p /tmp/bin
	@go build -o=/tmp/bin/$(binary_name) $(main_package_path)
	@$(call log_ok,... complete.)
	@echo
	
## build/linux_amd64: build the production binary -- OK
.PHONY: build/linux_amd64
build/linux_amd64:
	@$(call log_info,Building $(binary_name) for linux/amd64...)
	@mkdir -p /tmp/bin/linux_amd64
	@GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/$(binary_name) $(main_package_path)
	@$(call log_ok,... complete.)
	@echo
	
## run: run the application (optional ARGS passthrough)
.PHONY: run
run: env/check build
	@$(call log_info,Running $(binary_name)...)
	@$(ENV_LOAD); \
	/tmp/bin/$(binary_name) $(ARGS)
	
## run/live: run the application with reloading on file changes (optional ARGS passthrough)
.PHONY: run/live
run/live: env/check
	@$(call log_info,Running $(binary_name) with automatic refresh on file changes...)
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
		$(call log_err,ERROR: working tree is dirty (commit/stash first).); \
		echo "$$status" >&2; \
		exit 1; \
	fi

## require-upstream: fail unless current branch has an upstream tracking branch
.PHONY: require-upstream
require-upstream:
	@up="$$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true)"; \
	if [ -z "$$up" ]; then \
		$(call log_err,ERROR: no upstream set for this branch (run 'make push/u' first).); \
		exit 1; \
	fi
	
## on-main: fail unless currently on main
.PHONY: on-main
on-main:
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" != "main" ]; then \
		$(call log_err,ERROR: not on main (current: $$branch).); \
		exit 1; \
	fi

## on-feature: fail unless currently on a non-main branch with an approved prefix
.PHONY: on-feature
on-feature:
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" = "HEAD" ]; then \
		$(call log_err,ERROR: detached HEAD (checkout a branch).); \
		exit 1; \
	fi; \
	case "$$branch" in \
		feature/*|fix/*|refactor/*|chore/*|docs/*) ;; \
		main) \
			$(call log_err,ERROR: on main; use a work branch (feature/*, fix/*, refactor/*, chore/*, docs/*).); \
			exit 1 ;; \
		*) \
			$(call log_err,ERROR: branch '$$branch' not in allowed prefixes (feature/, fix/, refactor/, chore/, docs/).); \
			exit 1 ;; \
	esac

## up-to-date: fail unless local HEAD matches origin/main
.PHONY: up-to-date
up-to-date: on-main
	@git fetch origin; \
	if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse origin/main)" ]; then \
		$(call log_err,ERROR: local HEAD does not match origin/main (run 'make sync/main').); \
		exit 1; \
	fi

## repair/main: reset local main to origin/main (keeps a backup branch)
.PHONY: repair/main
repair/main: confirm require-clean on-main
	@$(call log_info,Repairing main...)
	@set -e; \
	git fetch origin; \
	backup="backup/main-local-$$(date +%Y%m%d-%H%M%S)"; \
	echo "Saving current main to '$$backup'..."; \
	git branch "$$backup" HEAD; \
	echo "Resetting main to origin/main..."; \
	git reset --hard origin/main
	@$(call log_ok,... complete.)
	@echo

## sync/main: fast-forward main from origin/main (no confirm; safe to call from other targets)
.PHONY: sync/main
sync/main: require-clean
	@$(call log_info,Syncing main from origin/main...)
	@git switch main >/dev/null
	@git pull --ff-only
	@$(call log_ok,... complete.)
	@echo

## sync/branch: rebase onto upstream then origin/main, audit, and publish (force-with-lease)
.PHONY: sync/branch
sync/branch: confirm require-clean on-feature require-upstream
	@$(call log_info,Syncing branch...)
	@$(call log_info,Rebasing branch onto upstream...)
	@git fetch origin
	@git rebase @{u}
	@$(call log_info,Rebasing branch onto origin/main...)
	@git rebase origin/main
	@$(call log_ok,... complete.)
	@echo
	@$(MAKE) --no-print-directory audit
	@git push --force-with-lease
	@$(call log_ok,... sync complete.)
	@echo

## sync: convenience alias for sync/branch
.PHONY: sync
sync: sync/branch

## rebase/upstream: rebase current branch onto its upstream (keeps branches linear)
.PHONY: rebase/upstream
rebase/upstream: confirm require-clean on-feature require-upstream
	@$(call log_info,Rebasing branch onto upstream...)
	@git fetch origin
	@git rebase @{u}
	@$(call log_ok,... complete.)
	@echo

## rebase/main: rebase current branch onto origin/main (keeps branches linear)
.PHONY: rebase/main
rebase/main: confirm require-clean on-feature
	@$(call log_info,Rebasing branch onto origin/main...)
	@git fetch origin
	@git rebase origin/main
	@$(call log_ok,... complete.)
	@echo

## branch/new: create and switch to a new work branch (from freshly synced main)
.PHONY: branch/new
branch/new: confirm require-clean
	@set -e; \
	$(MAKE) --no-print-directory on-main; \
	$(MAKE) --no-print-directory sync/main; \
	$(call log_info,Branch type \(feature|fix|refactor|chore|docs\):); \
	printf ">> " ; \
	read type ; \
	case "$$type" in feature|fix|refactor|chore|docs) ;; \
		*) $(call log_err,ERROR: invalid type '$$type'.); exit 1 ;; \
	esac ; \
	$(call log_info,Slug \(lowercase, digits, hyphens; e.g. add-login\):); \
	printf ">> " ; \
	read slug ; \
	case "$$slug" in ""|[^a-z0-9]*|*[!a-z0-9-]*) \
		$(call log_err,ERROR: invalid slug '$$slug'.); exit 1 ;; \
	esac ; \
	branch="$$type/$$slug" ; \
	$(call log_info,Creating branch $$branch from main...); \
	git switch -c "$$branch"
	
## branch/cleanup: delete the current local branch after syncing main (use only for branches without PRs)
.PHONY: branch/cleanup
branch/cleanup: confirm require-clean
	@set -e; \
	branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$branch" = "main" ]; then \
		$(call log_err,ERROR: refusing to delete 'main'.); \
		exit 1; \
	fi; \
	$(call log_info,Cleaning up branch '$$branch'...); \
	$(MAKE) --no-print-directory sync/main; \
	git branch -D "$$branch"
	@$(call log_ok,... complete.)
	@echo

## push: fast-forward-only push (cheap pushes); refuse if it would be non-fast-forward
.PHONY: push
push: on-feature require-upstream require-clean
	@git push

## push/u: push current branch and set upstream to origin
.PHONY: push/u
push/u: on-feature require-clean
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	$(call log_info,Pushing '$$branch' and setting upstream to origin...); \
	git push -u origin "$$branch"
	@$(call log_ok,... complete.)
	@echo
	
## pr/create: create a GitHub PR for the current branch
.PHONY: pr/create
pr/create: confirm audit on-feature
	@gh pr create --fill-verbose --editor

## pr/view: open the current PR in the browser
.PHONY: pr/view
pr/view: on-feature
	@$(call log_info,Opening PR in browser...)
	@gh pr view --web

## pr/merge: squash-merge the PR for the current branch
.PHONY: pr/merge
pr/merge: confirm require-clean on-feature
	@set -e; \
	branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	$(call log_info,Merging PR for branch '$$branch'...); \
	gh pr merge --squash --delete-branch
	@$(call log_ok,... complete.)
	@echo

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
	@command -v sqlc >/dev/null 2>&1  || { $(call die,Refusing: sqlc not found. Install it and try again.); }
	@command -v goose >/dev/null 2>&1 || { $(call die,Refusing: goose not found. Install it and try again.); }
	@command -v psql >/dev/null 2>&1  || { $(call die,Refusing: psql not found. Install Postgres client tools and try again.); }

## db/check: fail if required DB env vars are not set
.PHONY: db/check
db/check: env/check db/tools/check
	@$(ENV_LOAD); \
	test -n "$$DB_HOST"          || { $(call die,Refusing: DB_HOST is not set.); } ; \
	test -n "$$DB_PORT"          || { $(call die,Refusing: DB_PORT is not set.); } ; \
	test -n "$$DB_SSLMODE"       || { $(call die,Refusing: DB_SSLMODE is not set.); } ; \
	test -n "$$DB_NAME"          || { $(call die,Refusing: DB_NAME is not set.); } ; \
	test -n "$$DB_NAME_TEST"     || { $(call die,Refusing: DB_NAME_TEST is not set.); } ; \
	test -n "$$DB_USER_ADMIN"    || { $(call die,Refusing: DB_USER_ADMIN is not set.); } ; \
	test -n "$$DB_USER_MIGRATOR" || { $(call die,Refusing: DB_USER_MIGRATOR is not set.); } ; \
	test -n "$$DB_USER_APP"      || { $(call die,Refusing: DB_USER_APP is not set.); }

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

## db/migrate/new: create a new migration file (sequential numbering)
.PHONY: db/migrate/new
db/migrate/new: db/tools/check
	@test -d "$(migrations_dir)" || { $(call die,Refusing: migrations_dir not found: $(migrations_dir)); }
	@set -e; \
	$(call log_info,Migration name (e.g. add-users-table):); \
	printf ">> " ; \
	read name ; \
	if [ -z "$$name" ]; then \
		$(call die,Refusing: migration name is required.); \
	fi; \
	goose -s -dir "$(migrations_dir)" create "$$name" sql

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
	test -n "$$DB_NAME_TEST" || { $(call die,Refusing: DB_NAME_TEST is not set.); } ; \
	test "$$DB_NAME_TEST" != "$$DB_NAME" || { $(call die,Refusing: DB_NAME_TEST must not equal DB_NAME.); }

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
	@command -v upx >/dev/null 2>&1 || { $(call die,Refusing: upx not found. Install it or remove compression from deploy.); }

## predeploy: run checks required before a deployment
.PHONY: predeploy
predeploy: audit require-clean on-main up-to-date confirm
	@$(call log_ok,Pre-deploy checks passed.)
	@echo

## production/connect: connect to the production server
.PHONY: production/connect
production/connect: env/check
	@$(ENV_LOAD); \
	test -n "$$PRODUCTION_HOST_IP"   || { $(call die,Refusing: PRODUCTION_HOST_IP is not set.); } ; \
	test -n "$$PRODUCTION_SSH_USER"  || { $(call die,Refusing: PRODUCTION_SSH_USER is not set.); } ; \
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

