-- This file is used by SQLC to autogenerate the Go code needed by the database driver. 
-- It contains all tables required for storing user data needed by the Gateway.
-- See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries

/*--------------------------------------------------------------*/
/*-------------------- Enum Definitions ------------------------*/
/*--------------------------------------------------------------*/

-- Define the CapacityLimitPeriod enum
CREATE TYPE capacity_limit_period AS ENUM (
    'CAPACITY_LIMIT_PERIOD_UNSPECIFIED',
    'CAPACITY_LIMIT_PERIOD_DAILY',
    'CAPACITY_LIMIT_PERIOD_WEEKLY',
    'CAPACITY_LIMIT_PERIOD_MONTHLY'
);

-- Define the Auth_AuthType enum
CREATE TYPE auth_type AS ENUM (
    'NO_AUTH',
    'API_KEY_AUTH',
    'JWT_AUTH'
);

/*--------------------------------------------------------------*/
/*-------------------- Table Definitions -----------------------*/
/*--------------------------------------------------------------*/

-- Create the plans table
CREATE TABLE plans (
    name VARCHAR(255) PRIMARY KEY,
    throughput_limit INT NOT NULL,
    capacity_limit INT NOT NULL,
    capacity_limit_period capacity_limit_period NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create the gateway_endpoints table
CREATE TABLE gateway_endpoints (
    id VARCHAR(64) PRIMARY KEY,
    plan_name VARCHAR(255) NOT NULL REFERENCES plans(name),
    auth_type auth_type NOT NULL,
    api_key VARCHAR(255), -- Nullable, only used if auth_type is API_KEY_AUTH
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create the users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    auth_provider_user_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create the gateway_endpoint_users join table
CREATE TABLE gateway_endpoint_users (
    gateway_endpoint_id VARCHAR(64) NOT NULL REFERENCES gateway_endpoints(id),
    auth_provider_user_id VARCHAR(255) NOT NULL REFERENCES users(auth_provider_user_id),
    PRIMARY KEY (gateway_endpoint_id, auth_provider_user_id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

/*--------------------------------------------------------------*/
/*-------------------- Listener Updates ------------------------*/
/*--------------------------------------------------------------*/

-- Create the changes table with an 'is_delete' field
CREATE TABLE gateway_endpoint_changes (
    id SERIAL PRIMARY KEY,
    gateway_endpoint_id VARCHAR(64) NOT NULL REFERENCES gateway_endpoints(id),
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the trigger function with 'is_delete' handling
CREATE OR REPLACE FUNCTION log_gateway_endpoint_changes() RETURNS trigger AS $$
DECLARE
    endpoint_ids TEXT[];
    is_delete BOOLEAN := FALSE;
BEGIN
    endpoint_ids := ARRAY[]::TEXT[];

    IF TG_TABLE_NAME = 'gateway_endpoints' THEN
        IF TG_OP = 'DELETE' THEN
            is_delete := TRUE;
            endpoint_ids := array_append(endpoint_ids, OLD.id);
        ELSE
            endpoint_ids := array_append(endpoint_ids, NEW.id);
        END IF;

    ELSIF TG_TABLE_NAME = 'plans' THEN
        SELECT array_agg(ge.id) INTO endpoint_ids
        FROM gateway_endpoints ge
        WHERE ge.plan_name = COALESCE(NEW.name, OLD.name);

    ELSIF TG_TABLE_NAME = 'users' THEN
        SELECT array_agg(ge.id) INTO endpoint_ids
        FROM gateway_endpoints ge
        JOIN gateway_endpoint_users geu ON ge.id = geu.gateway_endpoint_id
        WHERE geu.auth_provider_user_id = COALESCE(NEW.auth_provider_user_id, OLD.auth_provider_user_id);

    ELSIF TG_TABLE_NAME = 'gateway_endpoint_users' THEN
        SELECT array_agg(ge.id) INTO endpoint_ids
        FROM gateway_endpoints ge
        WHERE ge.id = COALESCE(NEW.gateway_endpoint_id, OLD.gateway_endpoint_id);
    END IF;

    -- Remove duplicates
    SELECT ARRAY(SELECT DISTINCT unnest(endpoint_ids)) INTO endpoint_ids;

    -- Insert into changes table with 'is_delete' flag
    IF array_length(endpoint_ids, 1) > 0 THEN
        INSERT INTO gateway_endpoint_changes (gateway_endpoint_id, is_delete)
        SELECT unnest(endpoint_ids), is_delete;
    END IF;

    -- Send minimal notification
    PERFORM pg_notify('gateway_endpoint_changes', '');

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table

CREATE TRIGGER gateway_endpoints_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON gateway_endpoints
FOR EACH ROW EXECUTE FUNCTION log_gateway_endpoint_changes();

CREATE TRIGGER plans_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON plans
FOR EACH ROW EXECUTE FUNCTION log_gateway_endpoint_changes();

CREATE TRIGGER users_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION log_gateway_endpoint_changes();

CREATE TRIGGER gateway_endpoint_users_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON gateway_endpoint_users
FOR EACH ROW EXECUTE FUNCTION log_gateway_endpoint_changes();

/*--------------------------------------------------------------*/
/*-------------------- Automatic Timestamps --------------------*/
/*--------------------------------------------------------------*/

-- Create the trigger function for automatic timestamps
CREATE OR REPLACE FUNCTION update_timestamps() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    IF TG_OP = 'INSERT' THEN
        NEW.created_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic timestamps

CREATE TRIGGER update_plans_timestamps
BEFORE INSERT OR UPDATE ON plans
FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_gateway_endpoints_timestamps
BEFORE INSERT OR UPDATE ON gateway_endpoints
FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_users_timestamps
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_gateway_endpoint_users_timestamps
BEFORE INSERT OR UPDATE ON gateway_endpoint_users
FOR EACH ROW EXECUTE FUNCTION update_timestamps();