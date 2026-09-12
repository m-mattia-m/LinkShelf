-- password stays NOT NULL: an admin-invited (not-yet-completed) account
-- stores '' there, same convention CreateExternal already uses for
-- OIDC-provisioned accounts that have no local password either.
ALTER TABLE "user" ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE "user" ADD COLUMN verified_at TIMESTAMP;

CREATE TABLE IF NOT EXISTS "email_action_token" (
    id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    token_hash CHAR(64) NOT NULL,
    -- 'verify' (account already has a password, just confirming the
    -- address) or 'set_password' (admin-invited account completing
    -- registration by setting its first password).
    action VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_email_action_token PRIMARY KEY (id),
    CONSTRAINT uq_email_action_token_hash UNIQUE (token_hash),
    CONSTRAINT fk_email_action_token_user
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_email_action_token_user_id
    ON "email_action_token"(user_id);
