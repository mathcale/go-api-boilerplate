#!/bin/bash

if ! command -v migrate 2>&1 >/dev/null; then
  echo "'migrate' could not be found, please install it by running 'make install-deps'"
  exit 1
fi

read -p "==> Enter the migration name: " -r MIGRATION_NAME

migrate create -seq \
  -ext sql \
  -dir migrations \
  $MIGRATION_NAME
