package usecase

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

func TestGetConversionSettings(t *testing.T) {
	type args struct {
		db *sql.DB
		id int
	}
	tests := []struct {
		name      string
		args      args
		setupMock func(mock sqlmock.Sqlmock, args args)
		want      *ImageConversionSetting
		wantErr   bool
	}{
		{
			name: "valid case",
			args: args{id: 1},
			setupMock: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"output_format", "width_resize_ratio", "height_resize_ratio"}).
					AddRow("jpeg", 0.8, 0.8)
				mock.ExpectQuery(`SELECT output_format, width_resize_ratio, height_resize_ratio FROM image_conversion_settings WHERE id = \$1`).
					WithArgs(args.id).
					WillReturnRows(rows)
			},
			want: &ImageConversionSetting{
				OutputFormat: "jpeg",
				WidthResizeRatio: 0.8,
				HeightResizeRatio: 0.8,
			},
			wantErr: false,
		},
		{
			name: "invalid case",
			args: args{id: 2},
			setupMock: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT output_format, width_resize_ratio, height_resize_ratio FROM image_conversion_settings WHERE id = \$1`).
					WithArgs(args.id).
					WillReturnError(errors.New("DB error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.setupMock(mock, tt.args)
			got, err := GetConversionSettings(db, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConversionSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConversionSettings() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
