CREATE TABLE match_cats (
    id SERIAL PRIMARY KEY,
    issuedId INTEGER REFERENCES users(id) NOT NULL,
    issuedCatId INTEGER REFERENCES cats(id) NOT NULL,
    receiverId INTEGER REFERENCES users(id) NOT NULL,
    receiverCatId INTEGER REFERENCES cats(id) NOT NULL,
    message TEXT,
    status BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);