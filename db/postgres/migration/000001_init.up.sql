CREATE TYPE post_status AS ENUM(
    'DRAFT',        -- do not touch if updated_time < 24h * 3
    'SCHEDULED',    -- post wait publication
    'PUBLISHING',   -- post sended to Publisher
    'DONE',         -- post publish success
    'FAILED',       -- post publish go wrong
    'ARCHIVED'     -- post archived on 24h * 60
);

CREATE TABLE IF NOT EXISTS posts (
    "id" VARCHAR NOT NULL PRIMARY KEY,
    "publish_channel" VARCHAR NOT NULL,
    "data" JSONB NOT NULL,
    "status" post_status NOT NULL,
    "publish_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "updated_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "attempts" INT DEFAULT 0 NOT NULL CHECK("attempts" >= 0)
);

CREATE INDEX IF NOT EXISTS idx_posts_channel ON posts("publish_channel");
