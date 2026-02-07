# print this help message
[default]
help:
    @just --list

# runs audit & tests
[group('*workflow')]
check:
    @scripts/audit 
    @printf '\n'
    @scripts/test unit

# runs audit & tests -race
[group('*workflow')]
ci:
    scripts/audit
    scripts/test race

# runs tidy, gofmt, and go-mod-upgrade
[group('*workflow')]
maintain:
    scripts/tidy
    scripts/mod-upgrade

# runs a chain of QC commands
[group('quality')]
audit:
    scripts/audit

# runs tidy & gofmt
[group('quality')]
tidy:
    scripts/tidy

# runs go-mod-upgrade
[group('quality')]
mod-upgrade:
    scripts/mod-upgrade

# runs tests, accepts commands [unit|race|cover]
[group('test')]
test mode="":
    scripts/test {{mode}}

# builds command binary with native target
[group('build')]
build:
    scripts/build

# builds command binary with linux_amd64 target
[group('build')]
build-linux:
    scripts/build linux_amd64

# runs a built binary, accepts commands [default|test|live|debug]
[group('run')]
run mode="":
    scripts/run {{mode}}

# runs a built binary with live reload (air)
[group('run')]
run-live:
    scripts/run live

# fast-forward main from origin/main
[group('git')]
sync-main:
    scripts/git/sync-main

# rebase onto upstream then origin/main, audit, and publish (force-with-lease)
[group('git')]
sync-branch:
    scripts/git/sync-branch
