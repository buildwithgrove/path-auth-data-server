-- This file is used by SQLC to autogenerate the Go code needed by the Grove Portal database driver. 
-- It contains all tables required for storing user data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- For backwards compatibility, this file uses the tables defined in the Grove Portal database schema 
-- from the Portal HTTP DB (PHD) repo and is used in this repo to allow PATH to source its authorization
-- data from the existing Grove Portal Postgres database.
-- See: https://github.com/pokt-foundation/portal-http-db/blob/master/postgres-driver/sqlc/schema.sql

-- IMPORTANT - All tables and columns defined in this file exist in the existing Grove Portal DB.

-- The `portal_applications` and its associated tables are converted to the `proto.GatewayEndpoint` format.
-- The inline comments indicate the fields in the `proto.GatewayEndpoint` that correspond to the columns in the `portal_applications` table.

-- Plans Tables
CREATE TABLE pay_plans (
    plan_type VARCHAR(25) PRIMARY KEY,
    monthly_relay_limit INT NOT NULL, -- GatewayEndpoint.RateLimiting.CapacityLimit
    throughput_limit INT NOT NULL -- GatewayEndpoint.RateLimiting.ThroughputLimit
);

-- Accounts Tables
CREATE TABLE accounts (
    id VARCHAR(10) PRIMARY KEY, -- GatewayEndpoint.Metadata.AccountId
    plan_type VARCHAR(25) NOT NULL REFERENCES pay_plans(plan_type) -- GatewayEndpoint.Metadata.PlanType
);

-- Users Tables
CREATE TABLE users (
    id VARCHAR(10) PRIMARY KEY
);

-- User Auth Providers Tables
CREATE TYPE auth_type AS ENUM (
    'auth0_github',
    'auth0_username',
    'auth0_google'
);

CREATE TABLE user_auth_providers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type auth_type NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL, -- GatewayEndpoint.Auth.AuthType.Jwt.AuthorizedUsers[key]
    UNIQUE(user_id, type)
);

-- Account Users Tables
CREATE TABLE account_users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(10) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    UNIQUE (user_id, account_id)
);

-- Portal Application Tables
CREATE TABLE portal_applications (
    id VARCHAR(24) PRIMARY KEY UNIQUE, -- GatewayEndpoint.EndpointId
    account_id VARCHAR(10) REFERENCES accounts(id)
); 

-- Portal Application Settings Table
CREATE TABLE portal_application_settings (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(24) NOT NULL UNIQUE REFERENCES portal_applications(id) ON DELETE CASCADE,
    secret_key VARCHAR(64), -- GatewayEndpoint.Auth.AuthType.StaticApiKey.ApiKey
    secret_key_required BOOLEAN
);