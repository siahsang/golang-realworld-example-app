CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email    TEXT NOT NULL,
    password bytea NOT NULL
);
