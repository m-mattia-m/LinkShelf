-- Lets a shelf owner turn off the default "Powered by LinkShelf" public-page
-- footer, or replace it with their own short text (rendered as a restricted
-- subset of Markdown by the frontend - see ShelfPublicView). Existing shelves
-- default to footer_enabled = true with no custom text, which reproduces
-- today's only behavior (the default footer, unconditionally shown).
ALTER TABLE `shelf`
    ADD COLUMN footer_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN footer_custom_text VARCHAR(500) NOT NULL DEFAULT '';
