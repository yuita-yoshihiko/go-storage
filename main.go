package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"go-storage/infrastructure/db"
	"go-storage/usecase"
)

func main() {
	ctx := context.Background()
	cc := &usecase.MockClientCreator{}

	bucketName, id, err := validateAndParseArgs()
	if err != nil {
		log.Fatalf("引数のバリデーションに失敗しました: %v", err)
	}

	if err := processImage(ctx, cc, bucketName, id); err != nil {
		log.Fatalf("画像の処理に失敗しました: %v", err)
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

func processImage(ctx context.Context, cc usecase.ClientCreator, bucketName string, id int) error {
	client, err := usecase.CreateClient(ctx, cc)
	if err != nil {
		return fmt.Errorf("クライアントの作成に失敗しました: %w", err)
	}
	defer client.Close()

	latestObject, err := usecase.GetLatestObject(ctx, client, bucketName)
	if err != nil {
		return fmt.Errorf("画像の取得に失敗しました: %w", err)
	}

	dbInstance, err := db.CreateDBConnection()
	if err != nil {
		return fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}
	defer dbInstance.Close()

	shouldInsert, err := db.ShouldInsertData(dbInstance)
	if err != nil {
		return fmt.Errorf("初期データの挿入チェックに失敗しました: %w", err)
	}

	if shouldInsert {
		if err := db.InsertInitialData(dbInstance); err != nil {
			return fmt.Errorf("初期データの挿入に失敗しました: %w", err)
		}
	}

	objectName := latestObject.ObjectName()
	if err := db.SaveOriginalImageInfo(dbInstance, objectName); err != nil {
		return fmt.Errorf("元の画像データの保存に失敗しました: %w", err)
	}

	conversionSettings, err := usecase.GetConversionSettings(dbInstance, id)
	if err != nil {
		return fmt.Errorf("画像の変換設定の取得に失敗しました: %w", err)
	}

	data, err := usecase.DownloadImage(ctx, client, bucketName, objectName)
	if err != nil {
		return fmt.Errorf("画像のダウンロードに失敗しました: %w", err)
	}

	resizedData, err := usecase.ResizeImage(data, conversionSettings.WidthResizeRatio, conversionSettings.HeightResizeRatio, conversionSettings.OutputFormat)
	if err != nil {
		return fmt.Errorf("画像のリサイズに失敗しました: %w", err)
	}

	newObjectName := "resized_" + objectName
	if err := usecase.UploadImage(ctx, client, bucketName, newObjectName, resizedData); err != nil {
		return fmt.Errorf("画像のアップロードに失敗しました: %w", err)
	}

	if err := db.SaveResizedImageInfo(dbInstance, newObjectName); err != nil {
		return fmt.Errorf("リサイズ後の画像情報のデータベースへの保存に失敗しました: %w", err)
	}

	fmt.Printf("画像の処理に成功しました: %s\n", objectName)

	return nil
}
