SET TIME ZONE 'Asia/Ho_Chi_Minh';

ALTER DATABASE urlshortsevice SET timezone TO 'Asia/Ho_Chi_Minh';

CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(16) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    title TEXT,
    clicks BIGINT DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT (now()),
    expires_at timestamptz NULL
);

CREATE INDEX ON "urls" ("code");

CREATE INDEX ON "urls" ("original_url");