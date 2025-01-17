CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    level INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (name, description, level)
VALUES ('admin', 'Administrator: can update and delete posts of other users', 3),
       ('moderator', 'Moderator: can updated posts of other users', 2),
       ('user', 'Regular user: can create posts and comment', 1);