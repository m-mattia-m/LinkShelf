CREATE TABLE IF NOT EXISTS "oidc_state" (
    state VARCHAR(64) NOT NULL,
    code_verifier VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT pk_oidc_state PRIMARY KEY (state)
);
