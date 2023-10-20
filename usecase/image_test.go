package usecase

import (
	"io/ioutil"
	"testing"

	_ "github.com/lib/pq"
)

func TestCheckImageFormat(t *testing.T) {
	type args struct {
		data []byte
	}
	data, err := ioutil.ReadFile("../testdata/test.jpeg")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				data: data,
			},
			want:    "jpeg",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckImageFormat(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckImageFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckImageFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	type args struct {
		data   []byte
		width  float64
		height float64
		format string
	}
	data, err := ioutil.ReadFile("../testdata/test.jpeg")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				data:   data,
				width:  0.8,
				height: 0.8,
				format: "jpeg",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResizeImage(tt.args.data, tt.args.width, tt.args.height, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResizeImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
