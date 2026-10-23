#!/usr/bin/env sh

set -e

OPEN_REPORT=1
KEEP_PROFILE=0

for arg in "$@"; do
  case "$arg" in
    --no-open|-n)
      OPEN_REPORT=0
      ;;
    --keep)
      KEEP_PROFILE=1
      ;;
  esac
done

cleanup() {
  rm -f coverage.txt profile.out coverage.filtered.out

  if [ "$KEEP_PROFILE" -eq 0 ]; then
    rm -f coverage.tmp
  fi
}

trap cleanup EXIT

cleanup

go test -coverprofile=coverage.tmp -covermode=set ./...

grep -v -E -f .covignore coverage.tmp > coverage.filtered.out
mv coverage.filtered.out coverage.tmp

if [ "$OPEN_REPORT" -eq 1 ]; then
  go tool cover -html=coverage.tmp
else
  go tool cover -html=coverage.tmp -o coverage.html

  echo "Coverage report written to coverage.html"
fi
