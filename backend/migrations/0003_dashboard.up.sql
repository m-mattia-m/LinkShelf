ALTER TABLE "setting"
    ALTER COLUMN value TYPE TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_shelf_path
    ON "shelf"(path)
    WHERE path <> '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_shelf_domain
    ON "shelf"(domain)
    WHERE domain <> '';

INSERT INTO "user" (id, email, first_name, last_name, role, password)
VALUES ('018f1a3e-0000-7000-8000-000000000001', 'dev@linkshelf.local', 'Dev', 'User', 'user', 'mypassword')
ON CONFLICT (id) DO NOTHING;
