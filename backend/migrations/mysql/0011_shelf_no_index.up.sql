-- Lets a shelf owner opt their own shelf's public page out of search engine
-- indexing (a per-shelf "noindex", not an instance-wide one - the instance
-- itself, and every other shelf, stays crawlable). Existing shelves default
-- to false (crawlable) so nothing already indexed silently disappears.
ALTER TABLE `shelf`
    ADD COLUMN no_index BOOLEAN NOT NULL DEFAULT false;
