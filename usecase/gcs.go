package usecase

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/lib/pq"
	"google.golang.org/api/iterator"
)

type ClientCreator interface {
	NewClient(ctx context.Context) (*storage.Client, error)
}

type MockClientCreator struct {
	Client *storage.Client
	Err    error
}

func (m *MockClientCreator) NewClient(ctx context.Context) (*storage.Client, error) {
	return m.Client, m.Err
}

type StorageClient interface {
	Bucket(name string) *storage.BucketHandle
}

// 最新の画像を取得する
func GetLatestObject(ctx context.Context, client StorageClient, bucketName string) (*storage.ObjectHandle, error) {
	bkt := client.Bucket(bucketName)
	query := &storage.Query{}
	it := bkt.Objects(ctx, query)
	var latestObject *storage.ObjectHandle
	var latestTime time.Time
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		// ここで最新の画像を取得している
		if attrs.Updated.After(latestTime) {
			latestTime = attrs.Updated
			latestObject = bkt.Object(attrs.Name)
		}
	}
	if latestObject == nil {
		return nil, fmt.Errorf("画像が存在しません。")
	}
	return latestObject, nil
}

// Googleクライアントの作成（GCPを操作するための初期設定的な）
func CreateClient(ctx context.Context, cc ClientCreator) (*storage.Client, error) {
	return cc.NewClient(ctx)
}
