CREATE TABLE IF NOT EXISTS auth.auth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES auth.users(id)
        ON DELETE CASCADE,

    provider VARCHAR(50) NOT NULL
        CHECK (
            provider IN (
                'local',
                'google',
                'wallet'
            )
        ),

    provider_user_id VARCHAR(255),

    password_hash TEXT,

    wallet_address VARCHAR(255),

    access_token TEXT,
    refresh_token TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(provider, provider_user_id)
);