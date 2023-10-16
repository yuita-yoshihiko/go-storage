package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"go-storage/src"
)

func CreateDBConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s host=%s password=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PASSWORD"),
	)
	dbInstance, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return dbInstance, nil
}

func ShouldInsertData(dbInstance *sql.DB) (bool, error) {
	var count int
	err := dbInstance.QueryRow("SELECT COUNT(*) FROM image_conversion_settings").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func InsertInitialData(dbInstance *sql.DB) error {
	query := `INSERT INTO image_conversion_settings (output_format, resize_w, resize_h) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	settings := []src.ImageConversionSetting{
		{"jpg", 0.8, 0.8},
		{"png", 0.8, 0.8},
	}
	for _, setting := range settings {
		if _, err := dbInstance.Exec(query, setting.OutputFormat, setting.ResizeW, setting.ResizeH); err != nil {
			return err
		}
	}
	return nil
}

func SaveOriginalImageInfo(dbInstance *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO original_images (object_name) VALUES ($1)`
	_, err := dbInstance.Exec(query, gcsObjectName)
	return err
}

func SaveResizedImageInfo(dbInstance *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO converted_images (converted_object_name) VALUES ($1)`
	_, err := dbInstance.Exec(query, gcsObjectName)
	return err
}