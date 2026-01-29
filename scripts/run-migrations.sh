#!/bin/sh
# This script runs the migrations for the database using goose on container startup

/usr/local/bin/goose -dir /migrations postgres "$DB_URL" up