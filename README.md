**1. 編譯你的 GO 應用程序**

```
go mod tidy
```

```
go build -o bin/api ./cmd/api
```

- -o server：指定輸出檔案名為 server
- 給 Linux 用

```
GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api
```

- 可直接跑以下指令把 server run 起來

```
cp .env bin/.env
./bin/api
```

**2. 修改 Env 到對應的環境**

- 在正式環境中，應用通常依賴環境變數（如資料庫連線資訊、API 金鑰）。確保這些變數在正式環境中正確配置。

**3. 使用 Docker 部署**

- a. 編寫 Dockerfile：

```
   FROM golang:1.22.6 AS builder
   RUN     mkdir -p /go-api
   # 創建應用目錄
   WORKDIR /go-api
   # 複製所有檔案到容器
   COPY go.mod .
   COPY go.sum .
   # 下載依賴
   RUN go mod download
   COPY . .
   EXPOSE 8089
   # 構建應用並指定輸出二進位檔案的位置
   RUN go build -o bin/api ./cmd/api

   # 設置ENTRYPOINT為可執行檔案，啟動服務器
   ENTRYPOINT ["./bin/api"]
```

- b. 構建 Docker 鏡像：

```
   docker build -t 'go-api' .
```

- c. 運行 Docker 容器：

```
    docker run -d -p 8089:8089 go-api
```

也可用

```
    docker-compose up -d
```

## 💖 Support the Project

If this project saved you time or helped you build something cool, consider a one-time donation!

[![Liberapay One-Time](https://img.shields.io/badge/Liberapay-Donate%20Once-blue?style=for-the-badge&logo=liberapay)](https://liberapay.com/andre1502/donate)

| Amount                                                                          | Goal                     |
| :------------------------------------------------------------------------------ | :----------------------- |
| **[$5 USD](https://liberapay.com/andre1502/donate?amount=5.00&currency=USD)**   | ☕ Buy me a coffee       |
| **[$15 USD](https://liberapay.com/andre1502/donate?amount=15.00&currency=USD)** | 🍕 Buy me a pizza        |
| **[$50 USD](https://liberapay.com/andre1502/donate?amount=50.00&currency=USD)** | 🛠️ Support a new feature |
| **Custom**                                                                      | 🚀 Every bit helps!      |

> **Note:** 100% of your donation goes to development (excluding PayPal's standard transaction fee).
