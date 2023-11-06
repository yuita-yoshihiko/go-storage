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

func GetConversionSettings(db *sql.DB, id int) (*ImageConversionSetting, error) {
	query := `SELECT output_format, width_resize_ratio, height_resize_ratio FROM image_conversion_settings WHERE id = $1`
	setting := &ImageConversionSetting{}

	var format string
	var width, height float64
	if err := db.QueryRow(query, id).Scan(&format, &width, &height); err != nil {
		return nil, err
	}

	setting.OutputFormat = format
	setting.WidthResizeRatio = float64(width)
	setting.HeightResizeRatio = float64(height)

	return setting, nil
}
