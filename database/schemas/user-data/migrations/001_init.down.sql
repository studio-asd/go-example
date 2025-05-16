DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS users_pii;
DROP TABLE IF EXISTS user_secrets;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS security_roles;
DROP TABLE IF EXISTS security_role_permissions;
DROP TABLE IF EXISTS security_permissions;

DROP INDEX IF EXISTS idx_unq_us_external_id;
DROP INDEX IF EXISTS idx_unq_us_pii_email;
DROP INDEX IF EXISTS idx_unq_ussecrets_uid_sk_st;
DROP INDEX IF EXISTS idx_unq_ussecrets_external_id;
DROP INDEX IF EXISTS idx_ussecrets_uid_st;
DROP INDEX IF EXISTS idx_ussecrets_ver_sid;
DROP INDEX IF EXISTS idx_us_session_user_id;
DROP INDEX IF EXISTS idx_sec_perm_name;
DROP INDEX IF EXISTS idx_sec_perm_type_key;