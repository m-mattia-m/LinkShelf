CREATE TABLE IF NOT EXISTS `theme` (
    id CHAR(36) NOT NULL,
    scope VARCHAR(20) NOT NULL DEFAULT 'user',
    owner_user_id CHAR(36),
    name VARCHAR(255) NOT NULL,
    source_file VARCHAR(255),
    config TEXT NOT NULL,
    CONSTRAINT pk_theme PRIMARY KEY (id),
    CONSTRAINT fk_theme_owner_user
        FOREIGN KEY (owner_user_id)
        REFERENCES `user`(id)
        ON DELETE CASCADE,
    CONSTRAINT uq_theme_source_file UNIQUE (source_file)
);

CREATE INDEX idx_theme_owner_user_id
    ON `theme`(owner_user_id);

-- theme_id is deliberately NOT a foreign key: a shelf must keep remembering
-- which theme it had selected even after that theme row is deleted, so the
-- edit page can tell "never picked a theme" apart from "the one you picked
-- is gone" and prompt for a new choice. Existence is resolved and validated
-- at the application layer instead.
ALTER TABLE `shelf`
    ADD COLUMN theme_id CHAR(36);

CREATE INDEX idx_shelf_theme_id
    ON `shelf`(theme_id);

ALTER TABLE `shelf`
    DROP COLUMN theme;

ALTER TABLE `section`
    ADD COLUMN `order` INT NOT NULL DEFAULT 0;

-- Backfill existing rows with a stable order derived from their (time-ordered,
-- since ids are UUIDv7) id, rather than leaving every row tied at 0.
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY shelf_id ORDER BY id) - 1 AS rn
    FROM `section`
)
UPDATE `section` s
JOIN ranked ON s.id = ranked.id
SET s.`order` = ranked.rn;

ALTER TABLE `link`
    ADD COLUMN `order` INT NOT NULL DEFAULT 0;

WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY section_id ORDER BY id) - 1 AS rn
    FROM `link`
)
UPDATE `link` l
JOIN ranked ON l.id = ranked.id
SET l.`order` = ranked.rn;
