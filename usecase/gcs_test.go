package usecase

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// StorageClientのモックを作成
type mockStorageClient struct {
	objects []*storage.ObjectAttrs
}

// mockStorageClientにBucketメソッドを実装
func (m *mockStorageClient) Bucket(name string) *storage.BucketHandle {
	return &storage.BucketHandle{}
}

// mockBucketHandleはstorage.BucketHandleのモックです。
type mockBucketHandle struct {
	attrs map[string]*storage.ObjectAttrs
}

// mockBucketHandleにObjectメソッドを実装します。これはstorage.ObjectHandleを返します。
func (b *mockBucketHandle) Object(name string) *storage.ObjectHandle {
	return &storage.ObjectHandle{}
}

// Iteratorのモックを作成
type mockIterator struct {
	objects []*storage.ObjectAttrs
	index   int
}

// mockIteratorにNextメソッドを実装
func (m *mockIterator) Next() (*storage.ObjectAttrs, error) {
	if m.index >= len(m.objects) {
		return nil, iterator.Done
	}
	obj := m.objects[m.index]
	m.index++
	return obj, nil
}

// mockStorageClientにObjectsメソッドを実装
func (m *mockStorageClient) Objects(ctx context.Context, query *storage.Query) *storage.ObjectIterator {
	return &storage.ObjectIterator{}
}

// GetLatestObject関数とその他のユースケースロジック...

func TestGetLatestObject(t *testing.T) {
	now := time.Now()

	// テスト用のオブジェクト属性を作成
	objAttrs := []*storage.ObjectAttrs{
		{Name: "old.png", Updated: now.Add(-time.Hour)},
		{Name: "new.png", Updated: now},
	}

	// テストケース
	tests := []struct {
		name       string
		client     *mockStorageClient
		bucketName string
		want       string
		wantErr    bool
	}{
		{
			name:       "Successful retrieval of latest object",
			client:     &mockStorageClient{objects: objAttrs},
			bucketName: "test-bucket",
			want:       "new.png",
			wantErr:    false,
		},
		{
			name:       "Error when no objects present",
			client:     &mockStorageClient{objects: []*storage.ObjectAttrs{}},
			bucketName: "empty-bucket",
			want:       "",
			wantErr:    true,
		},
	}

	// テストケースを実行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLatestObject(context.Background(), tt.client, tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// オブジェクト名を確認する
			if !tt.wantErr {
				gotName := got.ObjectName() // ここで実際のObjectNameメソッドを呼び出す
				if gotName != tt.want {
					t.Errorf("GetLatestObject() = %v, want %v", gotName, tt.want)
				}
			}
		})
	}
}

func TestCreateClient(t *testing.T) {
	mockClient := &storage.Client{}
	var mockError error = nil

	mockCreator := &MockClientCreator{
		Client: mockClient,
		Err:    mockError,
	}

	ctx := context.Background()
	got, err := CreateClient(ctx, mockCreator)
	assert.NoError(t, err)
	assert.Equal(t, mockClient, got)
}
