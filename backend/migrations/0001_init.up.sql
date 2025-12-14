CREATE TABLE IF NOT EXISTS "user" (
    id CHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    CONSTRAINT pk_user PRIMARY KEY (id),
    CONSTRAINT uq_user_email UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS "shelf" (
    id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    path VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    description VARCHAR(255),
    theme VARCHAR(32),
    icon VARCHAR(255),
    user_id CHAR(36) NOT NULL,
    CONSTRAINT pk_shelf PRIMARY KEY (id),
    CONSTRAINT fk_shelf_user
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "section" (
    id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    shelf_id CHAR(36) NOT NULL,
    CONSTRAINT pk_section PRIMARY KEY (id),
    CONSTRAINT fk_section_shelf
        FOREIGN KEY (shelf_id)
        REFERENCES "shelf"(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "link" (
    id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL,
    icon VARCHAR(255) NOT NULL,
    color CHAR(7) DEFAULT '#000000',
    section_id CHAR(36) NOT NULL,
    CONSTRAINT pk_link PRIMARY KEY (id),
    CONSTRAINT fk_link_section
        FOREIGN KEY (section_id)
        REFERENCES "section"(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_shelf_user_id
    ON "shelf"(user_id);

CREATE INDEX IF NOT EXISTS idx_section_shelf_id
    ON "section"(shelf_id);

CREATE INDEX IF NOT EXISTS idx_link_section_id
    ON "link"(section_id);
