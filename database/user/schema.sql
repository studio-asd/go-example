-- drop tables.
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS user_secrets;
DROP TABLE IF EXISTS user_sessions;

DROP TYPE IF EXISTS secret_type;
CREATE TYPE loan_status AS ENUM('password', 'api_key');

-- tables and index.
CREATE TABLE IF NOT EXISTS users(
    id varchar PRIMARY KEY,
    user_name varchar(30) NOT NULL,
    user_email varchar NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS user_secrets(
    id varchar PRIMARY KEY,
    secret_type secret_type NOT NULL,
    secret_key varchar NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS user_sessions(
    id varchar PRIMARY KEY,
    user_id varchar NOT NULL,
    session_metadata jsonb NOT NULL,
    created_at timestamp NOT NULL,
    expired_at timestamp NOT NULL
);
