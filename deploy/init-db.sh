#!/bin/bash
set -euo pipefail

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE newapi;
    CREATE DATABASE logs;
    CREATE DATABASE sms;
EOSQL

for db in tokenjoy newapi logs sms; do
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$db" \
    -c "CREATE EXTENSION IF NOT EXISTS ltree;"
done

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d logs \
  -c "CREATE SCHEMA IF NOT EXISTS newapi;"
