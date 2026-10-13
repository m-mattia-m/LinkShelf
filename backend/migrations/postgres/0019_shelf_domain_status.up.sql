-- The status of a shelf's domain, written back by an external controller
-- (PUT /v1/service/domains/{domain}/status). Empty state = nothing reported yet.
-- Changing or clearing the domain resets all four columns.
ALTER TABLE "shelf" ADD COLUMN IF NOT EXISTS domain_state VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE "shelf" ADD COLUMN IF NOT EXISTS domain_reason VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE "shelf" ADD COLUMN IF NOT EXISTS domain_message VARCHAR(1024) NOT NULL DEFAULT '';
-- JSON array of {type, name, value}.
ALTER TABLE "shelf" ADD COLUMN IF NOT EXISTS domain_records VARCHAR(4096) NOT NULL DEFAULT '';
-- Unix seconds of the last domain change or status write.
ALTER TABLE "shelf" ADD COLUMN IF NOT EXISTS domain_updated_at BIGINT NOT NULL DEFAULT 0;

-- With customDomains.requireReady, two shelves may hold the same domain as long as
-- only one of them is ready. Uniqueness among all shelves (requireReady off) is
-- checked by the service, the database only guards the ready ones.
ALTER TABLE "shelf" DROP CONSTRAINT IF EXISTS uq_shelf_domain;
CREATE INDEX IF NOT EXISTS idx_shelf_domain ON "shelf" (domain);
CREATE UNIQUE INDEX IF NOT EXISTS uq_shelf_domain_ready ON "shelf" (domain) WHERE domain_state = 'ready';
