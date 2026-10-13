-- The status of a shelf's domain, written back by an external controller
-- (PUT /v1/service/domains/{domain}/status). Empty state = nothing reported yet.
-- Changing or clearing the domain resets all four columns.
ALTER TABLE `shelf` ADD COLUMN domain_state VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE `shelf` ADD COLUMN domain_reason VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE `shelf` ADD COLUMN domain_message VARCHAR(1024) NOT NULL DEFAULT '';
-- JSON array of {type, name, value}.
ALTER TABLE `shelf` ADD COLUMN domain_records VARCHAR(4096) NOT NULL DEFAULT '';
-- Unix seconds of the last domain change or status write.
ALTER TABLE `shelf` ADD COLUMN domain_updated_at BIGINT NOT NULL DEFAULT 0;

-- With customDomains.requireReady, two shelves may hold the same domain as long as
-- only one of them is ready. Uniqueness among all shelves (requireReady off) is
-- checked by the service, the database only guards the ready ones.
ALTER TABLE `shelf` DROP INDEX uq_shelf_domain;
CREATE INDEX idx_shelf_domain ON `shelf` (domain);
-- NULL unless ready, so the unique index only constrains the ready shelves.
ALTER TABLE `shelf` ADD COLUMN domain_ready VARCHAR(255) GENERATED ALWAYS AS (CASE WHEN domain_state = 'ready' THEN domain END) STORED;
CREATE UNIQUE INDEX uq_shelf_domain_ready ON `shelf` (domain_ready);
