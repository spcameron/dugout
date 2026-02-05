#!/usr/bin/env bash
# shellcheck shell=bash
#
# scripts/lib/paths.sh

set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)}"

source "$ROOT_DIR/scripts/lib/lib.sh"
source "$ROOT_DIR/scripts/lib/project.sh"

TMP_BASE="${TMPDIR:-/tmp}"
TMP_BASE="${TMP_BASE%/}"

BIN_DIR="${BIN_DIR:-$TMP_BASE/$BINARY_NAME/bin}"
BIN_PATH="${BIN_PATH:-$BIN_DIR/$BINARY_NAME}"
