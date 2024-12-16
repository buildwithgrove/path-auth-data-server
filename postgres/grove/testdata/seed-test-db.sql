-- This file updates the ephemeral Docker Postgres test database initialized in postgres/docker_test.go
-- with just enough data to run the test of the database driver using an actual Postgres DB instance.

-- Insert into the 'pay_plans' table
INSERT INTO pay_plans (
    plan_type,
    monthly_relay_limit,
    throughput_limit
)
VALUES ('PLAN_FREE', 1000, 30),
    ('PLAN_UNLIMITED', 0, 0);

-- Insert into the 'accounts' table
INSERT INTO accounts (id, plan_type)
VALUES ('account_1', 'PLAN_FREE'),
    ('account_2', 'PLAN_UNLIMITED'),
    ('account_3', 'PLAN_FREE');

-- Insert into the 'users' table
INSERT INTO users (id)
VALUES ('user_1'),
    ('user_2'),
    ('user_3');

-- Insert into the 'user_auth_providers' table
INSERT INTO user_auth_providers (user_id, type, provider_user_id)
VALUES ('user_1', 'auth0_username', 'provider_user_1'),
    ('user_2', 'auth0_username', 'provider_user_2'),
    ('user_3', 'auth0_username', 'provider_user_3');

-- Insert into the 'account_users' table
INSERT INTO account_users (user_id, account_id)
VALUES ('user_1', 'account_1'),
    ('user_2', 'account_2'),
    ('user_3', 'account_3');

-- Insert into the 'portal_applications' table
INSERT INTO portal_applications (id, account_id)
VALUES ('endpoint_1_no_auth', 'account_1'),
    ('endpoint_2_jwt', 'account_2'),
    ('endpoint_3_no_auth_static_key', 'account_3'),
    ('endpoint_4_no_auth', 'account_1'),
    ('endpoint_5_static_key', 'account_2');

-- Insert into the 'portal_application_settings' table
INSERT INTO portal_application_settings (application_id, secret_key_required, secret_key)
VALUES ('endpoint_1_no_auth', FALSE, NULL),
    ('endpoint_2_jwt', TRUE, 'secret_key_2'),
    ('endpoint_3_no_auth_static_key', TRUE, 'secret_key_3'),
    ('endpoint_4_no_auth', FALSE, NULL),
    ('endpoint_5_static_key', TRUE, 'secret_key_5');
