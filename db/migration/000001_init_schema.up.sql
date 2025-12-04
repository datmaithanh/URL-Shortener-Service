SET TIME ZONE 'Asia/Ho_Chi_Minh';

ALTER DATABASE urlshortsevice SET timezone TO 'Asia/Ho_Chi_Minh';

CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE,
    short_url TEXT UNIQUE, 
    original_url TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    clicks bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT (now()),
    expires_at timestamptz NOT NULL
);

CREATE INDEX ON "urls" ("code");

CREATE INDEX ON "urls" ("original_url");