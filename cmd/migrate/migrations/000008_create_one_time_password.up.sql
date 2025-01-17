CREATE TABLE IF NOT EXISTS one_time_passwords (
    token bytea PRIMARY KEY,
    user_id bigint NOT NULL,
    expired_at timestamp(0) with time zone DEFAULT NULL
)