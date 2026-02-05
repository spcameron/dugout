# print this help message
[default]
help:
    @just --list

# runs a chain of QC commands
[group('quality control')]
audit:
    @scripts/audit

# runs tests, accepts commands unit|race|cover
[group('quality control')]
test mode="unit":
    @scripts/test {{mode}}

# runs tidy & gofmt
[group('quality control')]
tidy:
    @scripts/tidy

# runs go-mod-upgrade
[group('quality control')]
mod-upgrade:
    @scripts/mod-upgrade

# runs audit & tests
[group('dev jobs')]
check:
    @scripts/audit 
    @scripts/test unit

# runs tidy, gofmt, and go-mod-upgrade
[group('dev jobs')]
maintain:
    @scripts/tidy
    @scripts/mod-upgrade
