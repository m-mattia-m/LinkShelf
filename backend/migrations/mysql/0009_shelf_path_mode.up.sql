-- Whether app.userBasedPaths was on when the shelf was created. A shelf's
-- public URL depends on that setting (/<path> or /<username>/<path>), so the
-- app compares this with the current setting to tell an owner that links they
-- shared before the setting changed no longer work.
--
-- Shelves that already exist were all created before the setting existed,
-- when every path was top-level, so false is the right value for them.
ALTER TABLE `shelf`
    ADD COLUMN created_user_based_paths BOOLEAN NOT NULL DEFAULT false;
