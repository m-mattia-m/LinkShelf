-- A shelf's domain is now stored trimmed and lowercased, without a trailing
-- dot or slash and without :80 or :443, since the public lookup compares it
-- with the Host header of the request. The API never validated it before, so
-- rows written earlier can differ in case or carry stray whitespace.
--
-- Only what is safe to do in SQL happens here. A value that is still not a
-- valid domain afterwards simply never matches a Host header, and editing the
-- shelf forces a valid one.
--
-- This fails on the unique constraint if two shelves had domains that only
-- differed in case. The error names the value, and one of the two shelves has
-- to be edited by hand before the migration can be run again.
UPDATE "shelf" SET domain = LOWER(TRIM(domain)) WHERE domain IS NOT NULL;
UPDATE "shelf" SET domain = NULL WHERE domain = '';
