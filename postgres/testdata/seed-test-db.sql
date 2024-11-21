-- This file updates the ephemeral Docker Postgres test database initialized in postgres/docker_test.go
-- with just enough data to run the test of the database driver using an actual Postgres DB instance.

-- Insert into the 'plans' table
INSERT INTO plans (
    name,
    throughput_limit,
    capacity_limit,
    capacity_limit_period
)
VALUES 
    ('PLAN_FREE', 30, 100000, 'CAPACITY_LIMIT_PERIOD_MONTHLY'),
    ('PLAN_UNLIMITED', 0, 0, 'CAPACITY_LIMIT_PERIOD_UNSPECIFIED');

-- Insert into the 'gateway_endpoints' table
INSERT INTO gateway_endpoints (id, plan_name, auth_type, api_key)
VALUES 
    ('endpoint_1', 'PLAN_UNLIMITED', 'JWT_AUTH', NULL),
    ('endpoint_2', 'PLAN_UNLIMITED', 'API_KEY_AUTH', 'secret_key_2'),
    ('endpoint_3', 'PLAN_FREE', 'NO_AUTH', NULL),
    ('endpoint_4', 'PLAN_FREE', 'API_KEY_AUTH', 'secret_key_4'),
    ('endpoint_5', 'PLAN_UNLIMITED', 'NO_AUTH', NULL);

-- Insert into the 'users' table
INSERT INTO users (auth_provider_user_id)
VALUES 
    ('provider_user_1'),
    ('provider_user_2'),
    ('provider_user_3');

-- Insert into the 'gateway_endpoint_users' table
INSERT INTO gateway_endpoint_users (gateway_endpoint_id, auth_provider_user_id)
VALUES 
    ('endpoint_1', 'provider_user_1'),
    ('endpoint_2', 'provider_user_2'),
    ('endpoint_3', 'provider_user_3'),
    ('endpoint_1', 'provider_user_2'),
    ('endpoint_1', 'provider_user_3');
