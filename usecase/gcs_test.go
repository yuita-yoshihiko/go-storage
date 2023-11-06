package usecase

import (
	"context"
	"reflect"
	"testing"

	"cloud.google.com/go/storage"
	_ "github.com/lib/pq"
)

func TestGetLatestObject(t *testing.T) {
	type args struct {
		ctx        context.Context
		client     StorageClient
		bucketName string
	}
	tests := []struct {
		name    string
		args    args
		want    *storage.ObjectHandle
		wantErr bool
	}{
		{
			name:    "valid case",
			args:    args{ctx: context.Background(), client: &StorageClientMock{}, bucketName: "test-bucket"},
			want:    &storage.ObjectHandle{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLatestObject(tt.args.ctx, tt.args.client, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatestObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateClient(t *testing.T) {
	type args struct {
		ctx context.Context
		cc  ClientCreator
	}
	tests := []struct {
		name    string
		args    args
		want    *storage.Client
		wantErr bool
	}{
		{
			name:    "valid case",
			args:    args{ctx: context.Background(), cc: &ClientCreatorMock{}},
			want:    &storage.Client{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateClient(tt.args.ctx, tt.args.cc)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
