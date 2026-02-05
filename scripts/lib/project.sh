#!/usr/bin/env bash
# shellcheck shell=bash
#
# scripts/lib/project.sh

set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)}"

source "$ROOT_DIR/scripts/lib/lib.sh"

CMD_DIR="${CMD_DIR:-$ROOT_DIR/cmd/dugout}"
BINARY_NAME="${BINARY_NAME:-$(basename "$CMD_DIR")}"

# Guardrails
[[ "$BINARY_NAME" != "." && "$BINARY_NAME" != "/" && -n "$BINARY_NAME" ]] || {
  die "Refusing: invalid BINARY_NAME derived from CMD_DIR: '$CMD_DIR'"
}
