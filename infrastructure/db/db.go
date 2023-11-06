package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"go-storage/usecase"
)

// DB接続を作成する
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

// 変換用データが存在するか確認する。存在する場合（countが0じゃない）はfalseを、存在しない場合（countが0）はtrueを返す
func ShouldInsertData(dbInstance *sql.DB) (bool, error) {
	var count int
	err := dbInstance.QueryRow("SELECT COUNT(*) FROM image_conversion_settings").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// 変換用データを挿入する
func InsertInitialData(dbInstance *sql.DB) error {
	query := `INSERT INTO image_conversion_settings (output_format, width_resize_ratio, height_resize_ratio) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	settings := []usecase.ImageConversionSetting{
		{"jpg", 0.8, 0.8},
		{"png", 0.8, 0.8},
	}
	for _, setting := range settings {
		if _, err := dbInstance.Exec(query, setting.OutputFormat, setting.WidthResizeRatio, setting.HeightResizeRatio); err != nil {
			return err
		}
	}
	return nil
}

// 元の画像の情報を保存する
func SaveOriginalImageInfo(dbInstance *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO original_images (object_name) VALUES ($1)`
	_, err := dbInstance.Exec(query, gcsObjectName)
	return err
}

// 変換後の画像の情報を保存する
func SaveResizedImageInfo(dbInstance *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO converted_images (converted_object_name) VALUES ($1)`
	_, err := dbInstance.Exec(query, gcsObjectName)
	return err
}
