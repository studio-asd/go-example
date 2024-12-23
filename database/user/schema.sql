-- drop tables.
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS user_secrets;

DROP TABLE IF EXISTS user_sessions;

DROP TABLE IF EXISTS user_roles;

DROP TABLE IF EXISTS security_roles;

DROP TABLE IF EXISTS security_role_permissions;

DROP TABLE IF EXISTS security_permissions;

DROP TYPE IF EXISTS secret_type;

CREATE TYPE loan_status AS ENUM ('password', 'api_key');

-- tables and index.
CREATE TABLE IF NOT EXISTS users (
    id varchar PRIMARY KEY,
    user_name varchar(30) NOT NULL,
    user_email varchar NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE user_pii (
    user_id VARCHAR PRIMARY KEY,
    phone_number VARCHAR NOT NULL,
    identity_number VARCHAR NOT NULL,
    identity_type INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS user_secrets (
    id varchar PRIMARY KEY,
    secret_type secret_type NOT NULL,
    secret_key varchar NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS user_sessions (
    id varchar PRIMARY KEY,
    user_id varchar NOT NULL,
    session_metadata jsonb NOT NULL,
    created_at timestamp NOT NULL,
    expired_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id varchar PRIMARY KEY,
    role_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    expired_at timestamp NOT NULL,
    UNIQUE (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS security_roles (
    role_id bigint GENERATED ALWAYS AS IDENTITY,
    role_name varchar NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

-- security_role_permissions maps role to permissions as one role can have more than one permission.
CREATE TABLE IF NOT EXISTS security_role_permissions (
    role_id bigint PRIMARY KEY,
    permission_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    -- prevent the role for having the same permissions.
    UNIQUE (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS security_permissions (
    permission_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    permission_name varchar NOT NULL,
    -- permission_type is the granular type of permission. For example, 'api_endpoint', 'file_access'.
    -- We don't want to use enum for the permission_type because we might want to add much more permission
    -- type in the future and adding more of them will changes to the enum.
    permission_type varchar NOT NULL,
    permission_key varchar NOT NULL,
    permission_value varchar NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);
