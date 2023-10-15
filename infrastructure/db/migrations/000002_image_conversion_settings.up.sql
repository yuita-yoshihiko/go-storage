CREATE TABLE image_conversion_settings (
    id SERIAL PRIMARY KEY,
    output_format VARCHAR(10),
    resize_w FLOAT,
    resize_h FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);