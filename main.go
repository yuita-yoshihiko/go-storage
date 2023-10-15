package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"go-storage/src"
	"go-storage/infrastructure/db"
)

func main() {
	ctx := context.Background()
	bucketName := os.Getenv("BUCKET_NAME")

	if len(os.Args) < 2 {
		fmt.Println("コマンドの後ろにIDを指定してください。 例: go run main.go 1")
		return
	}

	idArg := os.Args[1]
	id, err := strconv.Atoi(idArg)
	if err != nil {
		fmt.Printf("IDの値が不正です。")
		return
	}

	client, err := src.CreateClient(ctx)
	if err != nil {
		log.Fatalf("クライアントの作成に失敗しました: %v", err)
	}
	defer client.Close()

	latestObject, err := src.GetLatestObject(ctx, client, bucketName)
	if err != nil {
		log.Fatalf("画像の取得に失敗しました: %v", err)
	}

	dbInstance, err := db.CreateDBConnection()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}
	defer dbInstance.Close()

	if err := db.InsertInitialData(dbInstance); err != nil {
		log.Fatalf("初期データの挿入に失敗しました: %v", err)
	}

	objectName := latestObject.ObjectName()
	if err := db.SaveOriginalImageInfo(dbInstance, objectName); err != nil {
		log.Fatalf("元の画像データの保存に失敗しました。: %v", err)
	}

	conversionSettings, err := src.GetConversionSettings(dbInstance, id)
	if err != nil {
		log.Fatalf("画像の変換設定の取得に失敗しました: %v", err)
	}

	data, err := src.DownloadImage(ctx, client, bucketName, objectName)
	if err != nil {
		log.Fatalf("画像のダウンロードに失敗しました: %v", err)
	}

	resizedData, err := src.ResizeImage(data, conversionSettings.ResizeW, conversionSettings.ResizeH, conversionSettings.OutputFormat)
	if err != nil {
		log.Fatalf("画像のリサイズに失敗しました: %v", err)
	}

	newObjectName := "resized_" + objectName
	if err := src.UploadImage(ctx, client, bucketName, newObjectName, resizedData); err != nil {
		log.Fatalf("画像のアップロードに失敗しました: %v", err)
	}

	if err := db.SaveResizedImageInfo(dbInstance, newObjectName); err != nil {
		log.Fatalf("リサイズ後の画像情報のデータベースへの保存に失敗しました: %v", err)
	}

	fmt.Printf("画像の処理に成功しました。: %s\n", objectName)
}
