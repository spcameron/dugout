# print this help message
[default]
help:
    @just --list

# runs a chain of QC commands
[group('quality control')]
audit:
    scripts/audit

# runs tests, accepts commands unit|race|cover
[group('quality control')]
test mode="":
    scripts/test {{mode}}

# runs tidy & gofmt
[group('quality control')]
tidy:
    scripts/tidy

# runs go-mod-upgrade
[group('quality control')]
mod-upgrade:
    scripts/mod-upgrade

# runs audit & tests
[group('dev jobs')]
check:
    scripts/audit 
    scripts/test unit

# runs tidy, gofmt, and go-mod-upgrade
[group('dev jobs')]
maintain:
    scripts/tidy
    scripts/mod-upgrade

# builds command binary with native target
[group('dev jobs')]
build:
    scripts/build

# builds command binary with linux_amd64 target
[group('dev jobs')]
build-linux:
    scripts/build linux_amd64

# runs a built binary, modes [default|test|live|debug]
[group('dev jobs')]
run mode="":
    scripts/run {{mode}}
