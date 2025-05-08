-- This file updates the ephemeral Docker Postgres test database initialized in postgres/docker_test.go
-- with just enough data to run the test of the database driver using an actual Postgres DB instance.

-- Insert into the 'accounts' table
INSERT INTO accounts (id, plan_type)
VALUES ('account_1', 'PLAN_FREE'),
    ('account_2', 'PLAN_UNLIMITED'),
    ('account_3', 'PLAN_FREE');

-- Insert into the 'portal_applications' table
INSERT INTO portal_applications (id, account_id)
VALUES ('endpoint_1_no_auth', 'account_1'),
    ('endpoint_2_static_key', 'account_2'),
    ('endpoint_3_static_key', 'account_3'),
    ('endpoint_4_no_auth', 'account_1'),
    ('endpoint_5_static_key', 'account_2');

-- Insert into the 'portal_application_settings' table
INSERT INTO portal_application_settings (application_id, secret_key_required, secret_key)
VALUES ('endpoint_1_no_auth', FALSE, NULL),
    ('endpoint_2_static_key', TRUE, 'secret_key_2'),
    ('endpoint_3_static_key', TRUE, 'secret_key_3'),
    ('endpoint_4_no_auth', FALSE, NULL),
    ('endpoint_5_static_key', TRUE, 'secret_key_5');
