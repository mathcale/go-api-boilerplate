#!/bin/bash

if [ ! -f .env ]; then
  echo "Please create .env file"
  exit 1
fi

source .env

docker exec -e PGPASSWORD=$DATABASE_PASSWORD -it go-api-boilerplate-database psql -U $DATABASE_USER -w -c "DROP DATABASE IF EXISTS $DATABASE_NAME"

docker exec -e PGPASSWORD=$DATABASE_PASSWORD -it go-api-boilerplate-database psql -U $DATABASE_USER -w -c "CREATE DATABASE $DATABASE_NAME"
