package usecase

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type ImageConversionSetting struct {
	OutputFormat string
	ResizeW      float64
	ResizeH      float64
}

// 入力されたIDに応じてDBから画像変換設定を取得する
func GetConversionSettings(db *sql.DB, id int) (*ImageConversionSetting, error) {
	query := `SELECT output_format, resize_w, resize_h FROM image_conversion_settings WHERE id = $1`
	setting := &ImageConversionSetting{}

	var format string
	var width, height float64
	if err := db.QueryRow(query, id).Scan(&format, &width, &height); err != nil {
		return nil, err
	}

	setting.OutputFormat = format
	setting.ResizeW = float64(width)
	setting.ResizeH = float64(height)

	return setting, nil
}
