CREATE TABLE IF NOT EXISTS `oidc_state` (
    state VARCHAR(64) NOT NULL,
    code_verifier VARCHAR(255) NOT NULL,
    -- DATETIME, not TIMESTAMP: see 0004_auth.up.sql for why.
    expires_at DATETIME NOT NULL,
    CONSTRAINT pk_oidc_state PRIMARY KEY (state)
);
