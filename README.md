## 概要
GCSに保存された画像の変換処理を行うGoのバッチプログラムです。  

## 技術スタック
・Go(1.20)  
・postgresql  

## 共有事項  
・GCPのアカウントを作成し、**サービスアカウントキーの発行・GCSでのバケット作成は完了している**前提です。  
・バケットは1つだけ使用します。  
・JPEG(JPG)、PNGの画像のみ処理可能です。  
　それ以外の画像を処理しようとした場合や、画像データと拡張子が異なる場合(JPEG画像にPNGの拡張子など)も画像処理できません。  
・JPEG画像を処理する場合は実行コマンドの最後に1を、PNG画像を処理する場合は2を入力して実行してください(動作確認方法8を参照)。  
・処理対象の画像はGCSに最後に保存された画像です。(ユーザーがアップロードした画像を随時処理する想定だと、直近でアップロードされた画像は最後に保存されたものと一致するはず。それをローカル環境で再現するため。)

## 動作詳細  
1.コマンド実行後、処理対象の画像がダウンロードされDB(original_imagesテーブル)に保存される。  
2.処理可能な画像であれば縦横共に0.8倍にリサイズされる。  
3.リサイズされた画像はファイル名の先頭に **resized_** が足された上でGCSに再アップロードされ、DB(converted_imagesテーブル)にも保存される。  

## 動作確認方法
**1.当リポジトリをクローンする**  
```sh
git clone git@github.com:yuita-yoshihiko/go-storage.git
````  
もしくは  
```sh
git clone https://github.com/yuita-yoshihiko/go-storage.git
```  

**2.クローンしたリポジトリのディレクトリに移動**  
```sh
cd go-storage
```  

**3.環境変数が必要なため.envファイルを作成し.envsampleの値をコピー**  
```sh
touch .env
cp .envsample .env
```  

**4.バケット名を環境変数に書き込む(your_bucket_nameの部分は自身のGCSのバケット名を入力)**  
```sh
echo "BUCKET_NAME=your_bucket_name" >> .env
```  

**5.ルートディレクトリ配下にcredentials.jsonファイルを作成し、GCPのサービスアカウントキーの内容をコピー**  
```sh
touch credentials.json
```  

**6.Dockerイメージをビルドしコンテナを起動**  
```sh
docker-compose up -d --build
```  

**7.マイグレーションを実行**  
```sh
docker-compose exec app migrate -database "postgres://postgres:password@db:5432/postgres?sslmode=disable" -path ./infrastructure/db/migrations up
```  

**8.画像変換処理を実行($[Num}の部分はjpegの場合1、pngの場合2)**
```sh
docker-compose exec app go run main.go ${Num}
```  
