CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE refresh_sessions(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL, 
    revoked_at timestamptz,
    user_agent text,
    ip text, 
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refresh_sessions_user_id_idx ON refresh_sessions(user_id);
CREATE INDEX refresh_sessions_expires_at_idx ON refresh_sessions(expires_at);

CREATE TYPE demo_status AS enum (
    'available',
    'in_progress',
    'completed'
);

CREATE TABLE guests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    demo demo_status NOT NULL DEFAULT 'available',
    demo_start timestamptz,
    demo_end timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen timestamptz,

    CONSTRAINT demo_start_end_chk CHECK (
        (demo = 'available' AND demo_start IS NULL AND demo_end IS NULL) OR
        (demo = 'in_progress' AND demo_start IS NOT NULL AND demo_end IS NULL) OR
        (demo = 'completed' AND demo_start IS NOT NULL AND demo_end IS NOT NULL AND demo_end >= demo_start)
    )
);

CREATE TYPE dump_status AS enum (
    'new',
    'analyzed',
    'waiting_answers',
    'planned'
);

CREATE TABLE dumps (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES users(id),
    guest_id uuid REFERENCES guests(id) ON DELETE CASCADE,
    status dump_status NOT NULL DEFAULT 'new',

    raw_text text,
    raw_deleted_at timestamptz,
    raw_text_expires_at timestamptz,
    
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,

    CONSTRAINT dumps_owner_chk CHECK (
        (user_id IS NOT NULL AND guest_id IS NULL) OR (guest_id IS NOT NULL AND user_id IS NULL)
    )
);

CREATE OR REPLACE FUNCTION set_updated_at_if_changed()
RETURNS trigger AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at = now();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER dumps_updated_at_trigger
BEFORE UPDATE ON dumps
FOR EACH ROW
EXECUTE FUNCTION set_updated_at_if_changed();

CREATE INDEX dumps_user_created_idx ON dumps(user_id, created_at DESC);
CREATE INDEX dumps_guest_created_at_idx ON dumps(guest_id, created_at DESC);
CREATE INDEX dumps_raw_text_expires_at_idx ON dumps(raw_text_expires_at) WHERE raw_text_expires_at IS NOT NULL;

CREATE TYPE mood AS enum (
    'overwhelmed',
    'anxious',
    'tired',
    'neutral',
    'motivated'
);

CREATE TABLE dump_analysis (
    dump_id uuid PRIMARY KEY REFERENCES dumps(id) ON DELETE CASCADE,

    tasks_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    questions_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    mood mood,
    quote text,

    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE dump_answers (
    dump_id uuid PRIMARY KEY REFERENCES dumps(id) ON DELETE CASCADE,
    answers_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dump_id uuid NOT NULL REFERENCES dumps(id) ON DELETE CASCADE,

    title text NOT NULL DEFAULT 'Plan',
    created_at timestamptz NOT NULL DEFAULT now(),
    saved_at timestamptz,
    deleted_at timestamptz
);

CREATE INDEX plans_dump_id_created_at_idx ON plans(dump_id, created_at DESC);

CREATE TABLE plan_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id uuid REFERENCES plans(id) ON DELETE CASCADE,
    ord int NOT NULL,
    text text NOT NULL,
    done boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,

    CONSTRAINT plan_items_plan_ord_uniq UNIQUE (plan_id, ord)
);

CREATE INDEX plan_items_plan_id_idx ON plan_items(plan_id);