CREATE TABLE converted_images (
    id SERIAL PRIMARY KEY,
    converted_object_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);