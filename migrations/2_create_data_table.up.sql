BEGIN;

CREATE TABLE IF NOT EXISTS data_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('login_password', 'text', 'binary', 'bank_card')),
    data bytea NOT NULL,
    metadata text DEFAULT ''
);

COMMIT;