DROP DATABASE IF EXISTS gonews;
CREATE DATABASE gonews;

\c gonews;

DROP TABLE IF EXISTS posts, authors;

CREATE TABLE authors (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT REFERENCES authors(id) NOT NULL,
    title TEXT  NOT NULL,
    content TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    published_at BIGINT
);
