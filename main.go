package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"

	"cloud.google.com/go/storage"
	"github.com/nfnt/resize"
)

func createClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}

func downloadImage(ctx context.Context, client *storage.Client, bucketName, objectName string) ([]byte, error) {
	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}

func resizeImage(data []byte, width, height uint) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	resizedImg := resize.Resize(width, height, img, resize.Lanczos3)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, resizedImg, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func uploadImage(ctx context.Context, client *storage.Client, bucketName, objectName string, data []byte) error {
	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("uploadImage: unable to write data to bucket %q, file %q: %v", bucketName, objectName, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("uploadImage: unable to close writer for bucket %q, file %q: %v", bucketName, objectName, err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	bucketName := ""
	objectName := "" 

	client, err := createClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	data, err := downloadImage(ctx, client, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to download image: %v", err)
	}

	resizedData, err := resizeImage(data, 800, 600) 
	if err != nil {
		log.Fatalf("Failed to resize image: %v", err)
	}

	newObjectName := "resized_" + objectName
	err = uploadImage(ctx, client, bucketName, newObjectName, resizedData)
	if err != nil {
		log.Fatalf("Failed to upload image: %v", err)
	}

	fmt.Printf("successfully as %s.\n", newObjectName)
}