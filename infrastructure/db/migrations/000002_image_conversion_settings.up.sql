CREATE TABLE image_conversion_settings (
    id SERIAL PRIMARY KEY,
    output_format VARCHAR(10),
    width_resize_ratio FLOAT,
    height_resize_ratio FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);