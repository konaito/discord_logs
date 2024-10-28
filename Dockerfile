# Go言語のビルド環境を使用
FROM golang:1.20-alpine

# ワーキングディレクトリを設定
WORKDIR /app

# go.modとgo.sumを生成
RUN go mod init discordhooksapi && go mod tidy

# 必要なファイルをコピー
COPY . .

# Goアプリケーションをビルド
RUN go build -o discordhooksapi  # バイナリを /app 内に作成

# ポート8080を公開
EXPOSE 8080

# アプリケーションを実行
CMD ["/app/discordhooksapi"]
