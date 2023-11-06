package usecase

import (
	"io/ioutil"
	"testing"
)

const (
	testJPEGFilePath = "../testdata/test.jpeg"
	invalidData      = "invalid image"
	expectedFormat   = "jpeg"
)

func readTestData(t *testing.T, filepath string) []byte {
	t.Helper()
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}
	return data
}

func TestCheckImageFormat(t *testing.T) {
	data := readTestData(t, testJPEGFilePath)

	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "valid case",
			data: data,
			want: expectedFormat,
			wantErr : false,
		},
		{
			name: "invalid case",
			data: []byte(invalidData),
			want: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckImageFormat(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckImageFormat() error = %v, wantErr = %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("CheckImageFormat() = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	data := readTestData(t, testJPEGFilePath)

	tests := []struct {
		name    string
		data    []byte
		width   float64
		height  float64
		format  string
		wantErr bool
	}{
		{
			name: "valid case",
			data: data,
			width: 0.8,
			height: 0.8,
			format: expectedFormat,
			wantErr: false,
		},
		{
			name: "invalid case",
			data: []byte(invalidData),
			width: 0.8,
			height: 0.8,
			format: expectedFormat,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResizeImage(tt.data, tt.width, tt.height, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResizeImage() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
