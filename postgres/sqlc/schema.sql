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

/* -------------------------------------------------------------- */
/* -------------------- Table Definitions ----------------------- */
/* -------------------------------------------------------------- */

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

/* -------------------------------------------------------------- */
/* -------------------- Automatic Timestamps -------------------- */
/* -------------------------------------------------------------- */

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