-- 00001_enable_uuid_ossp.sql
-- Enables the uuid-ossp extension so later migrations can default UUID primary
-- keys with uuid_generate_v4(). Split from the table migration so the schema
-- source stays one-concern-per-file and the extension is versioned like any
-- other change.

-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose Down
DROP EXTENSION IF EXISTS "uuid-ossp";
