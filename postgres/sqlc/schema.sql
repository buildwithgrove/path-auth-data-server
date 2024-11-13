-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all tables required for storing user data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

-- It uses the tables defined in the Grove Portal database schema defined in the Portal HTTP DB (PHD) repo:
-- https://github.com/pokt-foundation/portal-http-db/blob/master/postgres-driver/sqlc/schema.sql

-- The `portal_applications` and its associated tables are converted to the `proto.GatewayEndpoint` format.

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

-- Portal Application Settings Table
CREATE TABLE portal_application_settings (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(24) NOT NULL UNIQUE REFERENCES portal_applications(id) ON DELETE CASCADE,
    secret_key_required BOOLEAN
);

-- /*-------------------- Listener Updates --------------------*/

-- Create the changes table with an 'is_delete' field
CREATE TABLE portal_application_changes (
    id SERIAL PRIMARY KEY,
    portal_app_id VARCHAR(24) NOT NULL,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the trigger function with 'is_delete' handling
CREATE OR REPLACE FUNCTION log_portal_application_changes() RETURNS trigger AS $$
DECLARE
    portal_app_ids TEXT[];
    is_delete BOOLEAN := FALSE;
BEGIN
    portal_app_ids := ARRAY[]::TEXT[];

    IF TG_TABLE_NAME = 'portal_applications' THEN
        IF TG_OP = 'DELETE' THEN
            is_delete := TRUE;
            portal_app_ids := array_append(portal_app_ids, OLD.id);
        ELSE
            portal_app_ids := array_append(portal_app_ids, NEW.id);
        END IF;

    ELSIF TG_TABLE_NAME = 'portal_application_settings' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        WHERE pa.id = COALESCE(NEW.application_id, OLD.application_id);

    ELSIF TG_TABLE_NAME = 'accounts' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        WHERE pa.account_id = COALESCE(NEW.id, OLD.id);

    ELSIF TG_TABLE_NAME = 'account_users' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        WHERE pa.account_id = COALESCE(NEW.account_id, OLD.account_id);

    ELSIF TG_TABLE_NAME = 'users' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        JOIN accounts a ON pa.account_id = a.id
        JOIN account_users au ON au.account_id = a.id
        WHERE au.user_id = COALESCE(NEW.id, OLD.id);

    ELSIF TG_TABLE_NAME = 'user_auth_providers' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        JOIN accounts a ON pa.account_id = a.id
        JOIN account_users au ON au.account_id = a.id
        WHERE au.user_id = COALESCE(NEW.user_id, OLD.user_id);

    ELSIF TG_TABLE_NAME = 'pay_plans' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        JOIN accounts a ON pa.account_id = a.id
        WHERE a.plan_type = COALESCE(NEW.plan_type, OLD.plan_type);
    END IF;

    -- Remove duplicates
    SELECT ARRAY(SELECT DISTINCT unnest(portal_app_ids)) INTO portal_app_ids;

    -- Insert into changes table with 'is_delete' flag
    IF array_length(portal_app_ids, 1) > 0 THEN
        INSERT INTO portal_application_changes (portal_app_id, is_delete)
        SELECT unnest(portal_app_ids), is_delete;
    END IF;

    -- Send minimal notification
    PERFORM pg_notify('portal_application_changes', '');

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table

CREATE TRIGGER portal_applications_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON portal_applications
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER portal_application_settings_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON portal_application_settings
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER accounts_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON accounts
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER account_users_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_users
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER users_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER user_auth_providers_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON user_auth_providers
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER pay_plans_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON pay_plans
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();
