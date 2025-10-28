# 使用 Go 1.25 base image
# Stage builder
FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

ARG TARGET_BUILD=api \
    REPO_NAME=go-api \
    BRANCH_NAME=dev \
    COMMIT_HASH=none \
    BUILD_DATE=none \
    BUILD_VERSION=none

# 設定工作目錄
WORKDIR /app

# 設定基礎時區
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 複製 go.mod、go.sum 並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製其餘程式碼
COPY . .

# 建立 binary（可選）
RUN --mount=type=cache,target=/go/pkg/mod go build \
  -ldflags "-X main.repoName=$REPO_NAME -X main.branchName=$BRANCH_NAME -X main.commitHash=$COMMIT_HASH -X main.buildDate=$BUILD_DATE -X main.version=$BUILD_VERSION -s -w" \
  -o ./bin/${TARGET_BUILD} ./cmd/${TARGET_BUILD}/main.go

# Stage final for api
FROM alpine:latest AS final-api

WORKDIR /app

RUN apk add --no-cache tzdata

# 設定基礎時區
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bin/api ./
RUN mkdir -p ./logs && chmod -R 777 ./logs

EXPOSE 8089

# 啟動應用
ENTRYPOINT ["./api"]

# Stage final for scheduler
FROM alpine:latest AS final-scheduler

WORKDIR /app

RUN apk add --no-cache tzdata

# 設定基礎時區
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bin/scheduler ./
RUN mkdir -p ./logs && chmod -R 777 ./logs

EXPOSE 8080

# 啟動應用
ENTRYPOINT ["./scheduler"]
