package usecase

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type ImageConversionSetting struct {
	OutputFormat string
	WidthResizeRatio      float64
	HeightResizeRatio      float64
}

// 入力されたIDに応じてDBから画像変換設定を取得する
func GetConversionSettings(db *sql.DB, id int) (*ImageConversionSetting, error) {
	query := `SELECT output_format, width_resize_ratio, height_resize_ratio FROM image_conversion_settings WHERE id = $1`
	setting := &ImageConversionSetting{}
	if err := db.QueryRow(query, id).Scan(&setting.OutputFormat, &setting.WidthResizeRatio, &setting.HeightResizeRatio); err != nil {
		return nil, err
	}
	return setting, nil
}
