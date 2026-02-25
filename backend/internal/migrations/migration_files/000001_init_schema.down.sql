DROP INDEX IF EXISTS plan_items_plan_id_idx;

DROP INDEX IF EXISTS plans_dump_id_created_at_idx;

DROP INDEX IF EXISTS dumps_raw_text_expires_at_idx;
DROP INDEX IF EXISTS dumps_guest_created_at_idx;
DROP INDEX IF EXISTS dumps_user_created_idx;

DROP INDEX IF EXISTS refresh_sessions_expires_at_idx;
DROP INDEX IF EXISTS refresh_sessions_user_id_idx;

DROP TRIGGER IF EXISTS dumps_updated_at_trigger ON dumps;
DROP FUNCTION IF EXISTS set_updated_at_if_changed();

DROP TABLE IF EXISTS plan_items;
DROP TABLE IF EXISTS plans;

DROP TABLE IF EXISTS dump_answers;
DROP TABLE IF EXISTS dump_analysis;

DROP TABLE IF EXISTS dumps;

DROP TABLE IF EXISTS refresh_sessions;

DROP TABLE IF EXISTS guests;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS mood;
DROP TYPE IF EXISTS dump_status;
DROP TYPE IF EXISTS demo_status;

DROP EXTENSION IF EXISTS pgcrypto;