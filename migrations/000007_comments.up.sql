CREATE TABLE IF NOT EXISTS comments
(
    id         SERIAL PRIMARY KEY,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    author_id  INTEGER     NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    article_id INTEGER     NOT NULL REFERENCES articles (id) ON DELETE CASCADE
);

