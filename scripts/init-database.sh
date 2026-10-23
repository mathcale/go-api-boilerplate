#!/usr/bin/env bash

if [ ! -f .env ]; then
  echo "Please create .env file"
  exit 1
fi

source .env

CONTAINER="go-api-boilerplate-database"

for _ in $(seq 1 30); do
  docker exec "$CONTAINER" pg_isready -U "$DATABASE_USER" -q && break
  sleep 1
done

docker exec -e PGPASSWORD=$DATABASE_PASSWORD "$CONTAINER" psql -U $DATABASE_USER -w -c "DROP DATABASE IF EXISTS $DATABASE_NAME"

docker exec -e PGPASSWORD=$DATABASE_PASSWORD "$CONTAINER" psql -U $DATABASE_USER -w -c "CREATE DATABASE $DATABASE_NAME"
