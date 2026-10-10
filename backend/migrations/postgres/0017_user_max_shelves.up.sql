-- NULL = unlimited.
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS max_shelves INTEGER CHECK (max_shelves >= 0);
