ALTER TABLE "user" ADD COLUMN provider VARCHAR(20) NOT NULL DEFAULT 'LOCAL';
ALTER TABLE "user" ADD COLUMN provider_id VARCHAR(255);
ALTER TABLE "user" ADD CONSTRAINT uq_user_provider_id UNIQUE (provider_id);

CREATE TABLE IF NOT EXISTS "refresh_token" (
    id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    token_hash CHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_refresh_token PRIMARY KEY (id),
    CONSTRAINT uq_refresh_token_hash UNIQUE (token_hash),
    CONSTRAINT fk_refresh_token_user
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_token_user_id
    ON "refresh_token"(user_id);
