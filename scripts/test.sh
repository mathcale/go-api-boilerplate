#!/bin/sh

set -e

open_report=1

for arg in "$@"; do
  case "$arg" in
    --no-open|-n)
      open_report=0
      ;;
  esac
done

cleanup() {
  rm -f coverage.txt coverage.tmp profile.out coverage.filtered.out
}

trap cleanup EXIT

cleanup

go test -coverprofile=coverage.tmp -covermode=set ./...

grep -v -E -f .covignore coverage.tmp > coverage.filtered.out
mv coverage.filtered.out coverage.tmp

if [ "$open_report" -eq 1 ]; then
  go tool cover -html=coverage.tmp
else
  go tool cover -html=coverage.tmp -o coverage.html
  echo "Coverage report written to coverage.html"
fi
