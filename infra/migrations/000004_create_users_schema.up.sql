CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email CITEXT UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,

    role VARCHAR(50) NOT NULL DEFAULT 'user'
        CHECK (role IN ('user', 'admin', 'seller','super_admin')),

    status VARCHAR(50) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'banned')),

    avatar_url VARCHAR(500),

    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_users_email
ON auth.users(email);

CREATE INDEX idx_auth_users_username
ON auth.users(username);

CREATE INDEX idx_auth_users_status
ON auth.users(status);