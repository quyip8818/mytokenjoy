#!/bin/bash
set -euo pipefail

# 创建 sms 用户
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
    CREATE USER sms WITH PASSWORD 'sms';
EOSQL

# 创建所有 database（apps 侧 owner=tokenjoy，sms 侧 owner=sms）
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
    -- apps 侧
    CREATE DATABASE newapi;
    CREATE DATABASE logs;

    -- sms 侧
    CREATE DATABASE sms OWNER sms;
    CREATE DATABASE sms_newapi OWNER sms;
    CREATE DATABASE sms_logs OWNER sms;
EOSQL

# 为各 database 安装所需 extensions
for db in tokenjoy newapi logs sms sms_newapi sms_logs; do
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "$db" <<-EOSQL
      CREATE EXTENSION IF NOT EXISTS ltree;
EOSQL
done

# 创建日志库的 schema
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "logs" <<-EOSQL
    CREATE SCHEMA IF NOT EXISTS newapi;
EOSQL

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "sms_logs" <<-EOSQL
    CREATE SCHEMA IF NOT EXISTS newapi AUTHORIZATION sms;
EOSQL
