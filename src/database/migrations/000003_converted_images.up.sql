CREATE TABLE converted_images (
    id SERIAL PRIMARY KEY,
    converted_object_name VARCHAR(255),
    error boolean NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);