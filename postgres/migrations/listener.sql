/* -------------------------------------------------------------- */
/* -------------------- Listener Updates ------------------------ */
/* -------------------------------------------------------------- */

-- Create the changes table with an 'is_delete' field
CREATE TABLE gateway_endpoint_changes (
    id SERIAL PRIMARY KEY,
    gateway_endpoint_id VARCHAR(64) NOT NULL,
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
