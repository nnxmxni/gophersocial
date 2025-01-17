CREATE TABLE IF NOT EXISTS users(
    id bigserial PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    pwd bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
)