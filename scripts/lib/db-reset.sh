#!/usr/bin/env bash
# Shared DB reset helpers for scripts/dev/reset.sh and scripts/dev-sms.sh.
# Source this file; do not execute directly.

PG_URL="postgres://tokenjoy:tokenjoy@127.0.0.1:5510/postgres"

# ponytail: use docker exec instead of local psql — avoids requiring psql on host.
# Upgrade path: if local psql is ever standardized, switch back.
_psql() {
  "${COMPOSE[@]}" exec -T postgres psql "$@"
}

# reset_apps_databases — drop/recreate tokenjoy, newapi, logs (apps side only)
reset_apps_databases() {
  _psql -U tokenjoy -d postgres -v ON_ERROR_STOP=1 <<-EOSQL
    -- Terminate lingering connections before DROP
    SELECT pg_terminate_backend(pid) FROM pg_stat_activity
      WHERE datname IN ('tokenjoy','newapi','logs') AND pid <> pg_backend_pid();
    DROP DATABASE IF EXISTS tokenjoy;
    DROP DATABASE IF EXISTS newapi;
    DROP DATABASE IF EXISTS logs;
    CREATE DATABASE tokenjoy;
    CREATE DATABASE newapi;
    CREATE DATABASE logs;
EOSQL

  for db in tokenjoy newapi logs; do
    _psql -U tokenjoy -d "$db" -c "CREATE EXTENSION IF NOT EXISTS ltree;"
  done

  _psql -U tokenjoy -d logs -c "CREATE SCHEMA IF NOT EXISTS newapi;"
}

# reset_sms_databases — drop/recreate sms, sms_newapi, sms_logs (sms side only)
reset_sms_databases() {
  _psql -U tokenjoy -d postgres -v ON_ERROR_STOP=1 <<-EOSQL
    SELECT pg_terminate_backend(pid) FROM pg_stat_activity
      WHERE datname IN ('sms','sms_newapi','sms_logs') AND pid <> pg_backend_pid();
    DROP DATABASE IF EXISTS sms;
    DROP DATABASE IF EXISTS sms_newapi;
    DROP DATABASE IF EXISTS sms_logs;
    CREATE DATABASE sms OWNER sms;
    CREATE DATABASE sms_newapi OWNER sms;
    CREATE DATABASE sms_logs OWNER sms;
EOSQL

  for db in sms sms_newapi sms_logs; do
    _psql -U tokenjoy -d "$db" -c "CREATE EXTENSION IF NOT EXISTS ltree;"
  done

  _psql -U tokenjoy -d sms_logs -c "CREATE SCHEMA IF NOT EXISTS newapi AUTHORIZATION sms;"
}
