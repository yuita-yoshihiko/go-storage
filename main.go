package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"go-storage/infrastructure/db"
	"go-storage/usecase"
	"cloud.google.com/go/storage"
)

func main() {
	ctx := context.Background()
	bucketName, id, err := validateAndParseArgs()
	if err != nil {
		log.Fatalf("引数のバリデーションに失敗しました")
	}
	if err := processImage(ctx, bucketName, id); err != nil {
		log.Fatalf("画像の処理に失敗しました")
	}
}

func validateAndParseArgs() (bucketName string, id int, err error) {
	bucketName = os.Getenv("BUCKET_NAME")
	if len(os.Args) < 2 {
		return "", 0, fmt.Errorf("コマンドの後ろにIDを指定してください。例: go run main.go 1")
	}
	idArg := os.Args[1]
	id, err = strconv.Atoi(idArg)
	if err != nil {
		return "", 0, fmt.Errorf("IDの値が不正です")
	}
	return bucketName, id, nil
}

func createClient(ctx context.Context) (*storage.Client, error) {
	client, err := usecase.CreateClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("クライアントの作成に失敗しました")
	}
	return client, nil
}

func connectToDatabase() (*sql.DB, error) {
	dbInstance, err := db.CreateDBConnection()
	if err != nil {
		return nil, fmt.Errorf("データベースへの接続に失敗しました")
	}
	return dbInstance, nil
}

func insertInitialData(dbInstance *sql.DB) error {
	shouldInsert, err := db.ShouldInsertData(dbInstance)
	if err != nil {
		return fmt.Errorf("画像リサイズ用データのチェックに失敗しました")
	}
	if shouldInsert {
		if err := db.InsertInitialData(dbInstance); err != nil {
			return fmt.Errorf("画像リサイズ用データの挿入に失敗しました")
		}
	}
	return nil
}

func processImage(ctx context.Context, bucketName string, id int) error {
	client, err := createClient(ctx)
	if err != nil {
		return err
	}
	dbInstance, err := connectToDatabase()
	if err != nil {
		return err
	}
	latestObject, err := usecase.GetLatestObject(ctx, client, bucketName)
	if err != nil {
		return fmt.Errorf("画像の取得に失敗しました")
	}
	objectName := latestObject.ObjectName()
	if err := db.SaveOriginalImageInfo(dbInstance, objectName); err != nil {
		return fmt.Errorf("元の画像データの保存に失敗しました")
	}
	data, err := usecase.DownloadImage(ctx, client, bucketName, objectName)
	if err != nil {
		return fmt.Errorf("画像のダウンロードに失敗しました")
	}
	conversionSettings, err := usecase.GetConversionSettings(dbInstance, id)
	if err != nil {
		return fmt.Errorf("画像の変換設定の取得に失敗しました")
	}
	resizedData, err := usecase.ResizeImage(data, conversionSettings.WidthResizeRatio, conversionSettings.HeightResizeRatio, conversionSettings.OutputFormat)
	if err != nil {
		return fmt.Errorf("画像のリサイズに失敗しました")
	}
	newObjectName := "resized_" + objectName
	if err := usecase.UploadImage(ctx, client, bucketName, newObjectName, resizedData); err != nil {
		return fmt.Errorf("画像のアップロードに失敗しました")
	}
	if err := db.SaveResizedImageInfo(dbInstance, newObjectName); err != nil {
		return fmt.Errorf("リサイズ後の画像情報のデータベースへの保存に失敗しました")
	}
	fmt.Printf("画像の処理に成功しました: %s\n", objectName)
	return nil
}
