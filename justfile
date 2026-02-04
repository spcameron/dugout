# print this help message
[default]
help:
    @just --list

# runs a chain of QC commands (fails fast)
[group('quality control')]
audit:
    @scripts/audit

# runs tests, accepts commands unit|race|cover
[group('quality control')]
test mode="unit":
    @scripts/test {{mode}}


# runs audit & tests (fails fast)
[group('quality control')]
check:
    @scripts/audit && scripts/test unit
