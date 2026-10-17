-- Complete development database rebuild entry point.
-- Run from the repository root with: psql -f sql/schema.sql
\set ON_ERROR_STOP on

-- Connect to a maintenance database because teachat itself will be dropped.
\connect postgres

DROP DATABASE IF EXISTS teachat WITH (FORCE);
CREATE DATABASE teachat OWNER robin;
\connect teachat

-- Build shared/core objects before schemas that reference them.
\ir schema_chat.sql
-- Development seed rows must exist before tea creates their accounts.
\ir seed_data.sql
\ir seed_industry_tags.sql
\ir schema_tea_payment_system.sql

-- Most DAO queries use unqualified core table names; tea queries are schema-qualified.
ALTER DATABASE teachat SET search_path TO chat, public;
SET search_path TO chat, public;