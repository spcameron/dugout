#!/usr/bin/env bash
# shellcheck shell=bash
#
# Lightweight logging & guard helpers for project scripts.
# Safe under: set -euo pipefail
#
# Usage:
#   source scripts/lib/log.sh
#   need_cmd git
#   require_file .env "Create .env from .env.example"
#   run go test ./...
#
# Optional:
#   LOG_TS=1     # prefix messages with [HH:MM:SS]
#   NO_COLOR=1   # disable ANSI color even when TTY

# ==================================================================================== #
# COLORS
# ==================================================================================== #

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
  BLUE=$'\033[0;34m'
  GREEN=$'\033[0;32m'
  YELLOW=$'\033[0;33m'
  RED=$'\033[0;31m'
  RESET=$'\033[0m'
else
  BLUE=''
  GREEN=''
  YELLOW=''
  RED=''
  RESET=''
fi

_ts() {
  [[ -n "${LOG_TS:-}" ]] && date +"%H:%M:%S"
}

_prefix() {
  local t
  t="$(_ts)"
  [[ -n "$t" ]] && printf '[%s] ' "$t"
}

info() {
  echo "${BLUE}$(_prefix)${RESET}$*"
}

ok() {
  echo "${GREEN}$(_prefix)✓ ${RESET}$*"
}

warn() {
  echo "${YELLOW}$(_prefix)! ${RESET}$*"
}

err() {
  echo "${RED}$(_prefix)✗ ${RESET}$*" >&2
}

die() {
  err "$*"
  exit 1
}

# ==================================================================================== #
# GUARDS
# ==================================================================================== #

need_cmd() { command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"; }

require_file() {
  # Usage: require_file "path" ["message"]
  local path="${1:?path required}"
  local msg="${2:-Expected file not found: $path}"

  [[ -f "$path" ]] || die "$msg"
}

require_dir() {
  # Usage: require_dir "path" ["message"]
  local path="${1:?path required}"
  local msg="${2:-Expected directory not found: $path}"

  [[ -d "$path" ]] || die "$msg"
}

require_env() {
  # Usage: require_env VAR ["message"]
  local var="${1:?env var name required}"
  local msg="${2:-Expected environment variable not set: $var}"

  [[ -n "${!var:-}" ]] || die "$msg"
}

# ==================================================================================== #
# CONVENIENCE
# ==================================================================================== #

run() {
  # Prints the command, then executes it.
  info "$*"
  "$@"
}
