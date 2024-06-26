CREATE TABLE cats (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    name VARCHAR(100) NOT NULL,
    race VARCHAR(50) CHECK (race IN ('Persian', 'Maine Coon', 'Siamese', 'Ragdoll', 'Bengal', 'Sphynx', 'British Shorthair', 'Abyssinian', 'Scottish Fold', 'Birman')) NOT NULL,
    sex VARCHAR(10) CHECK (sex IN ('male', 'female')) NOT NULL,
    age_in_month INTEGER,
    description TEXT,
    image_urls TEXT[],
    has_matched BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);