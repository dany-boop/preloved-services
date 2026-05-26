CREATE TABLE IF NOT EXISTS users.user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES auth.users(id)
        ON DELETE CASCADE,

    full_name VARCHAR(255),

    bio TEXT,

    location VARCHAR(255),

    website VARCHAR(500),

    phone VARCHAR(50),

    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id)
);

CREATE INDEX idx_user_profiles_user_id
ON users.user_profiles(user_id);