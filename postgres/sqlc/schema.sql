-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all tables required for storing user data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- Plans Tables
CREATE TABLE pay_plans (
    plan_type VARCHAR(25) PRIMARY KEY,
    monthly_relay_limit INT NOT NULL, -- GatewayEndpoint.RateLimiting.ThroughputLimit
    throughput_limit INT NOT NULL -- GatewayEndpoint.RateLimiting.CapacityLimit
);

-- Accounts Tables
CREATE TABLE accounts (
    id VARCHAR(10) PRIMARY KEY, -- GatewayEndpoint.UserAccount.AccountId
    plan_type VARCHAR(25) NOT NULL REFERENCES pay_plans(plan_type) -- GatewayEndpoint.UserAccount.PlanType
);

-- Users Tables
CREATE TABLE users (
    id VARCHAR(10) PRIMARY KEY
);

-- User Auth Providers Tables
CREATE TYPE auth_type AS ENUM ('auth0_github', 'auth0_username', 'auth0_google');
CREATE TABLE user_auth_providers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type auth_type NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL, -- GatewayEndpoint.Auth.AuthorizedUsers[key]
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