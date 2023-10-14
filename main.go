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
	"database/sql"
	"time"
	"os"
	"strings"
	"errors"
	"image/png"
	"strconv"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"github.com/nfnt/resize"
	_ "github.com/lib/pq"
)

type ImageConversionSetting struct {
	OutputFormat string
	ResizeW      float64
	ResizeH      float64
}

func createDBConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s host=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_HOST"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getLatestObject(ctx context.Context, client *storage.Client, bucketName string) (*storage.ObjectHandle, error) {
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

func insertInitialData(db *sql.DB) error {
	query := `INSERT INTO image_conversion_settings (output_format, resize_w, resize_h) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	settings := []ImageConversionSetting{
		{"jpeg", 0.8, 0.8},
	}
	for _, setting := range settings {
		if _, err := db.Exec(query, setting.OutputFormat, setting.ResizeW, setting.ResizeH); err != nil {
			return err
		}
	}
	return nil
}

func getConversionSettings(db *sql.DB, id int) (*ImageConversionSetting, error) {
	query := `SELECT output_format, resize_w, resize_h FROM image_conversion_settings WHERE id = $1`
	setting := &ImageConversionSetting{}

	var format string
	var width, height float64
	if err := db.QueryRow(query, id).Scan(&format, &width, &height); err != nil {
		return nil, err
	}

	setting.OutputFormat = format
	setting.ResizeW = float64(width)
	setting.ResizeH = float64(height)

	return setting, nil
}

func saveOriginalImageInfo(db *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO original_images (object_name) VALUES ($1)`
	_, err := db.Exec(query, gcsObjectName)
	return err
}

func saveResizedImageInfo(db *sql.DB, gcsObjectName string) error {
	query := `INSERT INTO converted_images (converted_object_name) VALUES ($1)`
	_, err := db.Exec(query, gcsObjectName)
	return err
}

func createClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}

func downloadImage(ctx context.Context, client *storage.Client, bucketName, objectName string) ([]byte, error) {
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

	format, err := checkImageFormat(data)
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

func checkImageFormat(data []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	if format != "jpeg" && format != "png" {
		return "", errors.New("jpeg(jpg)、またはpng形式の画像を選択してください。")
	}

	return format, nil
}

func resizeImage(data []byte, width, height float64, format string) ([]byte, error) {
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

func uploadImage(ctx context.Context, client *storage.Client, bucketName, objectName string, data []byte) error {
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

func main() {
	ctx := context.Background()
	bucketName := os.Getenv("BUCKET_NAME")

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [ID]")
		return
	}

	idArg := os.Args[1]
	id, err := strconv.Atoi(idArg)
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", idArg)
		return
	}

	client, err := createClient(ctx)
	if err != nil {
		log.Fatalf("クライアントの作成に失敗しました: %v", err)
	}
	defer client.Close()

	latestObject, err := getLatestObject(ctx, client, bucketName)
	if err != nil {
		log.Fatalf("画像の取得に失敗しました: %v", err)
	}

	db, err := createDBConnection()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}
	defer db.Close()

	if err := insertInitialData(db); err != nil {
		log.Fatalf("初期データの挿入に失敗しました: %v", err)
	}

	objectName := latestObject.ObjectName()
	if err := saveOriginalImageInfo(db, objectName); err != nil {
		log.Fatalf("元の画像データの保存に失敗しました。: %v", err)
	}

	conversionSettings, err := getConversionSettings(db, id)
	if err != nil {
		log.Fatalf("画像の変換設定の取得に失敗しました: %v", err)
	}

	data, err := downloadImage(ctx, client, bucketName, objectName)
	if err != nil {
		log.Fatalf("画像のダウンロードに失敗しました: %v", err)
	}

	resizedData, err := resizeImage(data, conversionSettings.ResizeW, conversionSettings.ResizeH, conversionSettings.OutputFormat)
	if err != nil {
		log.Fatalf("画像のリサイズに失敗しました: %v", err)
	}

	newObjectName := "resized_" + objectName
	if err := uploadImage(ctx, client, bucketName, newObjectName, resizedData); err != nil {
		log.Fatalf("画像のアップロードに失敗しました: %v", err)
	}

	if err := saveResizedImageInfo(db, newObjectName); err != nil {
		log.Fatalf("リサイズ後の画像情報のデータベースへの保存に失敗しました: %v", err)
	}

	fmt.Printf("画像の処理に成功しました。: %s\n", objectName)
}
