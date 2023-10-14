CREATE TABLE original_images (
    id SERIAL PRIMARY KEY,
    object_name VARCHAR(255) NOT NULL,
    error boolean NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);