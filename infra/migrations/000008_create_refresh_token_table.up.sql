CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES auth.users(id)
        ON DELETE CASCADE,

    token VARCHAR(500) UNIQUE NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ,

    ip_address INET,

    user_agent VARCHAR(500),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token
ON auth.refresh_tokens(token);

CREATE INDEX idx_refresh_tokens_user_id
ON auth.refresh_tokens(user_id);

CREATE INDEX idx_active_refresh_tokens
ON auth.refresh_tokens(user_id)
WHERE revoked_at IS NULL;