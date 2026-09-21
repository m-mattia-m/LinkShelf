-- username is NULL for accounts created before it existed. The backend fills
-- it in on startup (see domain.BackfillUsernames), and every creation path
-- sets it from then on, so NULL only survives until the next start. It is
-- stored lowercase, which makes a plain UNIQUE constraint case-insensitive.
ALTER TABLE `user`
    ADD COLUMN username VARCHAR(30) NULL;

ALTER TABLE `user`
    ADD CONSTRAINT uq_user_username UNIQUE (username);

-- With app.userBasedPaths a shelf is addressed as /<username>/<path>, so a
-- path only has to be unique per owner. The config can change at any time,
-- which a migration cannot follow, so the constraint is always the looser
-- per-owner one. The application enforces instance-wide uniqueness itself
-- while the feature is off, and refuses to start if switching it off would
-- leave two shelves with the same path.
ALTER TABLE `shelf`
    ADD CONSTRAINT uq_shelf_user_path UNIQUE (user_id, path);

ALTER TABLE `shelf`
    DROP INDEX uq_shelf_path;
