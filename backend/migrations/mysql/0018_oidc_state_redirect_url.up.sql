-- Empty = the configured authentication.oidc.redirectUrl.
ALTER TABLE `oidc_state` ADD COLUMN redirect_url VARCHAR(2048) NOT NULL DEFAULT '';
