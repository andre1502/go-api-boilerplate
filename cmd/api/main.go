package main

import (
	"context"
	"fmt"
	"go-api-boilerplate/cmd"
	"go-api-boilerplate/cmd/api/server"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

var (
	repoName   string
	branchName string
	commitHash string
	buildDate  string
	version    string
)

var isShuttingDown int32

func main() {
	container := cmd.MainInit(repoName, branchName, commitHash, buildDate, version, &isShuttingDown)

	// 註冊路由器
	container.Provide(server.NewServer)

	// 啟動服務
	err := container.Invoke(func(server *server.Server, cfg *config.Config) {
		fmt.Println("logger initialized")
		logger.Log.Info("logger initialized")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			fmt.Println("start Server...")
			logger.Log.Info("start Server...")

			sc := echo.StartConfig{
				Address:    cfg.SERVER_ADDRESS,
				HideBanner: true,
				// HidePort:        true,
				GracefulTimeout: 60 * time.Second,
				OnShutdownError: func(err error) {
					fmt.Printf("shutdown Server error %v\n", err)
					logger.Log.Errorf("shutdown Server error %v\n", err)
				},
			}

			if err := sc.Start(ctx, server.Echo); err != nil {
				fmt.Printf("Start Server error %v\n", err)
				logger.Log.Errorf("Start Server error %v\n", err)
			}
		}()

		<-ctx.Done()

		fmt.Println("server exiting.")
		logger.Log.Info("server exiting.")
	})

	if err != nil {
		fmt.Printf("container.Invoke Server failed: %v\n", err)
		logger.Log.Errorf("container.Invoke Server failed: %v", err)

		os.Exit(1)
		return
	}

	os.Exit(0)
}
