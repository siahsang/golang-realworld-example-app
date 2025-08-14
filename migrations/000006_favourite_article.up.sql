CREATE TABLE IF NOT EXISTS favourite_articles
(
    article_id INTEGER NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, user_id)
);
