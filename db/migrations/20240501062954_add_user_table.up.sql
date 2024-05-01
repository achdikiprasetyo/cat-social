CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    token_expired_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);