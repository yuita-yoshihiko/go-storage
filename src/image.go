package src

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"strings"
	
	"cloud.google.com/go/storage"
	"github.com/nfnt/resize"
	_ "github.com/lib/pq"
)

func DownloadImage(ctx context.Context, client *storage.Client, bucketName, objectName string) ([]byte, error) {
	if !strings.HasSuffix(strings.ToLower(objectName), ".jpg") && !strings.HasSuffix(strings.ToLower(objectName), ".jpeg") && !strings.HasSuffix(strings.ToLower(objectName), ".png") {
		return nil, errors.New("jpeg(jpg)、またはpng形式の画像を選択してください。")
	}

	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	format, err := CheckImageFormat(data)
	if err != nil {
		return nil, err
	}

	if format == "jpeg" && !strings.HasSuffix(strings.ToLower(objectName), ".jpg") && !strings.HasSuffix(strings.ToLower(objectName), ".jpeg") {
		return nil, errors.New("ファイルの拡張子と実際の画像タイプが異なっています。")
	} else if format == "png" && !strings.HasSuffix(strings.ToLower(objectName), ".png") {
		return nil, errors.New("ファイルの拡張子と実際の画像タイプが異なっています。")
	}

	return data, nil
}

func CheckImageFormat(data []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	if format != "jpeg" && format != "png" {
		return "", errors.New("jpeg(jpg)、またはpng形式の画像を選択してください。")
	}

	return format, nil
}

func ResizeImage(data []byte, width, height float64, format string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	newWidth := uint(float64(img.Bounds().Dx()) * width)
	newHeight := uint(float64(img.Bounds().Dy()) * height)
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	buf := new(bytes.Buffer)

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, resizedImg, nil)
	case "png":
		err = png.Encode(buf, resizedImg)
	default:
		return nil, fmt.Errorf("変換後の画像タイプがうまく取得できませんでした。: %s", format)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UploadImage(ctx context.Context, client *storage.Client, bucketName, objectName string, data []byte) error {
	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("画像のアップロードに失敗しました。: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("バケットへの書き込みがクローズできません。: %v", err)
	}

	return nil
}
