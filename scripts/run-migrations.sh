#!/bin/sh
# This script runs the migrations for the database using goose on container startup

if [ -f /.env ]; then
  set -a
  . /.env
  set +a
fi

DB_URL=${DB_URL:-postgres://postgres:postgres-pass@localhost:5432/smart?sslmode=disable}

/usr/local/bin/goose -dir /migrations postgres "$DB_URL" up