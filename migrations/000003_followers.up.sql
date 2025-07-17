CREATE TABLE IF NOT EXISTS followers
(
    user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    follower_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    primary key (user_id, follower_id)
);
