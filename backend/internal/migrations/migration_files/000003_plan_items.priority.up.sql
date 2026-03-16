CREATE TYPE priority AS enum (
    'low',
    'medium',
    'high'
);

ALTER TABLE plan_items
ADD COLUMN priority priority DEFAULT 'low';