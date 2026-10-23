#!/usr/bin/env bash

OLD_NAME="go-api-boilerplate"
OLD_DISPLAY_NAME="Go API Boilerplate"

read -p "==> Enter your project's module name (e.g. my-api): " -r PROJECT_NAME
read -p "==> Enter your project's display name (e.g. My API): " -r DISPLAY_NAME

grep -rlE "$OLD_NAME|$OLD_DISPLAY_NAME" . \
  --exclude-dir=.git \
  --exclude=rename-pkgs.sh \
  --exclude=README.md |
  xargs sed -i \
    -e "s|$OLD_NAME|$PROJECT_NAME|g" \
    -e "s|$OLD_DISPLAY_NAME|$DISPLAY_NAME|g"
