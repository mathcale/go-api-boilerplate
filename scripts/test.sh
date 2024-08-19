#!/bin/sh

set -e

cleanup() {
  rm -f coverage.txt coverage.tmp profile.out
}

trap cleanup EXIT

cleanup
echo "mode: set" > coverage.tmp

for d in $(go list ./... | grep -v vendor); do
  go test -coverprofile=profile.out -covermode=set $d

  if [ -f profile.out ]; then
    tail -n +2 profile.out >>coverage.tmp
    rm profile.out
  fi
done

cat coverage.tmp | grep -v "/internal/tests/mocks" > coverage.txt

go tool cover -html=coverage.txt
