CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE ,
    email    TEXT NOT NULL UNIQUE ,
    password bytea NOT NULL
);

