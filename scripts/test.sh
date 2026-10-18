#!/bin/sh

set -e

open_report=1
keep_profile=0

for arg in "$@"; do
  case "$arg" in
    --no-open|-n)
      open_report=0
      ;;
    --keep)
      keep_profile=1
      ;;
  esac
done

cleanup() {
  rm -f coverage.txt profile.out coverage.filtered.out
  if [ "$keep_profile" -eq 0 ]; then
    rm -f coverage.tmp
  fi
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
