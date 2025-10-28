package main

import (
	"context"
	"fmt"
	"go-api-boilerplate/cmd"
	"go-api-boilerplate/cmd/api/server"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

		fmt.Println("start Server...")
		logger.Log.Info("start Server...")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			if err := server.Echo.Start(cfg.SERVER_ADDRESS); err != http.ErrServerClosed {
				fmt.Printf("start Server error %v\n", err)
				logger.Log.Errorf("start Server error %v\n", err)
			}
		}()

		<-ctx.Done()

		// Wait for interrupt signal to gracefully shutdown the server with a timeout of 60 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := server.Echo.Shutdown(ctx); err != nil {
			fmt.Printf("shutdown Server error %v\n", err)
			logger.Log.Errorf("shutdown Server error %v\n", err)
		}

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
