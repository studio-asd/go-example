DROP SCHEMA IF EXISTS user_data;

CREATE SCHEMA IF NOT EXISTS user_data;

DROP TABLE IF EXISTS user_data.users;

DROP TABLE IF EXISTS user_data.users_pii;

DROP TABLE IF EXISTS user_data.user_secrets;

DROP TABLE IF EXISTS user_data.user_sessions;

DROP TABLE IF EXISTS user_data.user_roles;

DROP TABLE IF EXISTS user_data.security_roles;

DROP TABLE IF EXISTS user_data.security_role_permissions;

DROP TABLE IF EXISTS user_data.security_permissions;

CREATE TABLE IF NOT EXISTS user_data.users (
    user_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- external_id is used as unique identifier for the user in external API.
    -- We use uuid_v4 to generate the external_id.
    external_id varchar NOT NULL,
    user_email varchar NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE user_data.users_pii (
    user_id bigint PRIMARY KEY,
    phone_number VARCHAR NOT NULL,
    identity_number VARCHAR NOT NULL,
    identity_type INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS user_data.user_secrets (
    secret_id bigint generated always as identity primary key,
    -- external_id is the id that used to identify a secret from the client side.
    external_id varchar NOT NULL,
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

CREATE TABLE IF NOT EXISTS user_data.user_secret_versions (
    secret_id bigint PRIMARY KEY,
    secret_version bigint NOT NULL,
    secret_value varchar NOT NULL,
    created_at timestamptz NOT NULL 
);

CREATE TABLE IF NOT EXISTS user_data.user_sessions (
    user_id bigint NOT NULL,
    -- random_number is a pure random number generated to give the uniqueness to the session
    -- identifier as we use user_id, random_number and created_time to identify the session.
    -- The idea is to form a base64(user_id, random_number, created_time) to form a token and
    -- use it as a session identifier on the client side.
    random_number int NOT NULL,
    created_time bigint NOT NULL,
    created_from_ip inet NOT NULL,
    -- created_from_loc tracks from where the session is created if available.
    created_from_loc varchar,
    created_from_user_agent varchar NOT NULL,
    session_metadata jsonb NOT NULL,
    expired_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, random_number, created_time)
);

CREATE TABLE IF NOT EXISTS user_data.user_roles (
    user_id bigint PRIMARY KEY,
    role_id bigint NOT NULL,
    created_at timestamptz NOT NULL,
    expired_at timestamptz NOT NULL,
    UNIQUE (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS user_data.security_roles (
    role_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_name varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- security_role_permissions maps role to permissions as one role can have more than one permission.
CREATE TABLE IF NOT EXISTS user_data.security_role_permissions (
    role_id bigint PRIMARY KEY,
    permission_id bigint NOT NULL,
    created_at timestamptz NOT NULL,
    -- prevent the role for having the same permissions.
    UNIQUE (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_data.security_permissions (
    permission_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    permission_name varchar NOT NULL,
    -- permission_type is the granular type of permission. For example, 'api_endpoint', 'file_access'.
    -- We don't want to use enum for the permission_type because we might want to add much more permission
    -- type in the future and adding more of them will changes to the enum.
    permission_type varchar NOT NULL,
    permission_key varchar NOT NULL,
    permission_value varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
