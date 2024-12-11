ALTER TABLE
    users
ADD COLUMN email_verified_at timestamp(0) with time zone DEFAULT NULL;