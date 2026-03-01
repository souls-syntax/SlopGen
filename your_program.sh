#!/bin/sh
set -e

(
  cd "$(dirname "$0")" # Ensure compile steps are run within the repository directory
  go build -o /tmp/slopGen app/*.go
)

exec /tmp/slopGen "$@"
