package src

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	_ "github.com/lib/pq"
)

type StorageClient interface {
	Bucket(name string) *storage.BucketHandle
}

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

func CreateClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}