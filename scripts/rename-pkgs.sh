#!/bin/bash

read -p "==> Enter your project's name (e.g. mathcale/go-api-boilerplate): " -r PROJECT_NAME

grep --exclude=rename-pkgs.sh --exclude-dir=.git "mathcale/go-api-boilerplate" . -lr | xargs sed -i "s|mathcale/go-api-boilerplate|$PROJECT_NAME|g"
