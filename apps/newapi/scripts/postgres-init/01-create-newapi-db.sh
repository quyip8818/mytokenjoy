#!/bin/bash
set -euo pipefail

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
	CREATE EXTENSION IF NOT EXISTS ltree;
	CREATE DATABASE newapi;
	CREATE DATABASE logs;
EOSQL
