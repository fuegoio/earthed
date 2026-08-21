#!/bin/sh
# Create the relay database alongside the default earthed database.
# PostgreSQL runs scripts in /docker-entrypoint-initdb.d/ on first init only.
set -e
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres <<-EOSQL
    CREATE DATABASE earthed_relay;
EOSQL
