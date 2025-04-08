CREATE TABLE IF NOT EXISTS users (
    user_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- external_id is used as unique identifier for the user in external API.
    -- We use uuid_v4 to generate the external_id.
    external_id varchar NOT NULL,
    security_roles varchar[] NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unq_us_external_id ON users("external_id");
