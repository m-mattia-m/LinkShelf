ALTER TABLE "setting"
    ALTER COLUMN value TYPE TEXT;

-- path/domain are stored as NULL (not '') when unset, so a plain UNIQUE
-- constraint enforces "unique when set, any number of shelves may be unset" -
-- this shape works identically on MySQL, which has no partial-index support.
ALTER TABLE "shelf"
    ALTER COLUMN path DROP NOT NULL;

UPDATE "shelf" SET path = NULL WHERE path = '';
UPDATE "shelf" SET domain = NULL WHERE domain = '';

ALTER TABLE "shelf"
    ADD CONSTRAINT uq_shelf_path UNIQUE (path);

ALTER TABLE "shelf"
    ADD CONSTRAINT uq_shelf_domain UNIQUE (domain);

INSERT INTO "user" (id, email, first_name, last_name, role, password)
VALUES ('018f1a3e-0000-7000-8000-000000000001', 'dev@linkshelf.local', 'Dev', 'User', 'user', 'mypassword')
ON CONFLICT (id) DO NOTHING;
