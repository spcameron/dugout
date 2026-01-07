-- db/bootstrap.sql
-- Create roles and databases for local development.
-- Run via: make db/bootstrap
--
-- Notes:
-- - This file intentionally does NOT set passwords.
-- - Re-running is safe-ish:
--     * roles are created conditionally (no error if they exist)
--     * CREATE DATABASE is attempted and may error "already exists"
--       (the Makefile treats those specific errors as acceptable)
-- - Ownership is set to the migrator role; the app role gets CONNECT only.
-- - Schema/table privileges are handled by migrations.
--
-- 1) Roles
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT
            1
        FROM
            pg_roles
        WHERE
            rolname = 'dugout_app') THEN
    CREATE ROLE dugout_app LOGIN;
END IF;
    IF NOT EXISTS (
        SELECT
            1
        FROM
            pg_roles
        WHERE
            rolname = 'dugout_migrator') THEN
    CREATE ROLE dugout_migrator LOGIN;
END IF;
END
$$;

-- 2) Databases
CREATE DATABASE dugout_dev OWNER dugout_migrator;

CREATE DATABASE dugout_test OWNER dugout_migrator;

-- 3) Ensure ownership (covers "DB already existed" case)
ALTER DATABASE dugout_dev OWNER TO dugout_migrator;

ALTER DATABASE dugout_test OWNER TO dugout_migrator;

-- 4) Allow connections (and lock down PUBLIC)
REVOKE CONNECT ON DATABASE dugout_dev FROM PUBLIC;

REVOKE CONNECT ON DATABASE dugout_test FROM PUBLIC;

GRANT CONNECT ON DATABASE dugout_dev TO dugout_app, dugout_migrator;

GRANT CONNECT ON DATABASE dugout_test TO dugout_app, dugout_migrator;

