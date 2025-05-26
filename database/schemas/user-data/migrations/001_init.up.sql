CREATE TABLE IF NOT EXISTS users (
    user_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- user_uuid is used as unique identifier for the user in external API.
    -- We use uuid_v4.
    user_uuid uuid NOT NULL,
    security_roles varchar[] NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unq_us_uuid ON users("user_uuid");

CREATE TABLE user_pii (
    user_id bigint PRIMARY KEY,
    email varchar NOT NULL,
    phone_number varchar,
    identity_number varchar,
    identity_type int,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unq_us_pii_email ON user_pii("email");

CREATE TABLE IF NOT EXISTS user_secrets (
    secret_id bigint generated always as identity primary key,
    secret_uuid uuid NOT NULL,
    user_id bigint NOT NULL,
    -- secret_key is a key identifier for the user so its easier for them to identify
    -- what the purpose of the secret is.
    -- An example of the secret_key is "user_password".
    secret_key varchar NOT NULL,
    secret_type int NOT NULL,
    current_secret_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    -- The secret key is unique per user and type.
    UNIQUE(user_id, secret_key, secret_type)
);

-- This index is used to ensure all secret_key is unique per user and secret type. Other than that the index is also useful to retrieve a specific
-- secret by user_id, secret_key and secret_type, for example in the login scenario because we already know the secret_key and secret_type.
CREATE UNIQUE INDEX IF NOT EXISTS idx_unq_ussecrets_uid_sk_st ON user_secrets("user_id", "secret_key", "secret_type");
-- This index is used to retrieve all secrets for a user under a specific secret type.
CREATE INDEX IF NOT EXISTS idx_ussecrets_uid_st ON user_secrets("user_id", "secret_type");

CREATE TABLE IF NOT EXISTS user_secret_versions (
    secret_id bigint NOT NULL,
    secret_version bigint NOT NULL,
    secret_value varchar NOT NULL,
    secret_salt varchar,
    created_at timestamptz NOT NULL ,
    PRIMARY KEY(secret_id, secret_version)
);

-- This index is used to retrieve all secret versions for a specific secret id.
CREATE INDEX IF NOT EXISTS idx_ussecrets_ver_sid ON user_secrets("secret_id");

CREATE TABLE IF NOT EXISTS user_sessions (
    session_id uuid PRIMARY KEY,
    -- previous_sesision_id is used to track the previous session id if available. As we are allowing
    -- guess, the guess might create a new session as an authenticated user. Otherwise it will be an
    -- authenticated user that creates a new session.
    previous_sesision_id uuid,
    -- session_type is used to track the type of session. For example, 'authenticated', 'guest'.
    session_type int NOT NULL,
    -- user_id is the user that creates the session. The user_id can be NULL in case of guest session.
    -- The question might be, what if the user_id is NULL while the session_type is 'authenticated'?
    -- This kind of thing can happen because of bug in the program, and because there is no further validation
    -- in the database, it is possible to happen. In this case, the session will be invalid even though the
    -- client can generate the correct session_id.
    user_id bigint,
    -- random_id is a pure random number generated to give the uniqueness to the session
    -- identifier as we use user_id, random_id and created_at to identify the session.
    -- The idea is to form a base64(user_id, random_id, created_at) to form a token and
    -- use it as a session identifier on the client side.
    random_id VARCHAR NOT NULL,
    -- retrieving the end user IP address is sometimes tricky, but we will still store it anyway
    -- as tracking the IP is imporant for the session.
    created_from_ip inet NOT NULL,
    -- created_from_loc tracks from where the session is created if available.
    created_from_loc varchar,
    created_from_user_agent varchar NOT NULL,
    -- session_metadata stores information for the session. The field can be null in case of guest session.
    session_metadata jsonb,
    created_at timestamptz NOT NULL,
    expired_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_us_session_user_id ON user_sessions("user_id") WHERE user_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS security_roles (
    role_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_uuid uuid NOT NULL,
    role_name varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

-- security_role_permissions maps role to permissions as one role can have more than one permission.
CREATE TABLE IF NOT EXISTS security_role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at timestamptz NOT NULL,
    -- prevent the role for having the same permissions.
    PRIMARY KEY(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_sec_role_perm_role_id ON security_role_permissions("role_id");

CREATE TABLE IF NOT EXISTS security_permissions (
    permission_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    permission_uuid uuid NOT NULL,
    permission_name varchar NOT NULL,
    -- permission_type is the granular type of permission. For example, 'api_endpoint', 'file_access'.
    -- We don't want to use enum for the permission_type because we might want to add much more permission
    -- type in the future and adding more of them will changes to the enum.
    permission_type varchar NOT NULL,
    permission_key varchar NOT NULL,
    permission_value varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sec_perm_name  ON security_permissions("permission_name");
CREATE UNIQUE INDEX IF NOT EXISTS idx_sec_perm_type_key  ON security_permissions("permission_type", "permission_key");
