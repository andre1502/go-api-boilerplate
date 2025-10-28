package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-api-boilerplate/cmd"
	"go-api-boilerplate/internal/scheduler"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
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

	// Scheduler
	container.Provide(scheduler.NewScheduler)

	// 啟動服務
	err := container.Invoke(func(scheduler *scheduler.Scheduler, cfg *config.Config) {
		fmt.Println("logger initialized")
		logger.Log.Info("logger initialized")

		fmt.Println("start Scheduler...")
		logger.Log.Info("start Scheduler...")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			scheduler.Cron.Start()
			fmt.Println("cron start.")
			logger.Log.Info("cron start.")

			startHealthServer(ctx, cfg)
		}()

		<-ctx.Done()

		atomic.StoreInt32(&isShuttingDown, 1)

		fmt.Println("Start to shutdown Scheduler...")
		logger.Log.Info("Start to shutdown Scheduler...")

		if err := scheduler.Cron.Shutdown(); err != nil {
			fmt.Printf("Shutdown Scheduler error %v\n", err)
			logger.Log.Errorf("Shutdown Scheduler error %v\n", err)
		}

		fmt.Println("Scheduler shutdown...")
		logger.Log.Info("Scheduler shutdown...")

		// Wait for interrupt signal to gracefully shutdown the server with a timeout of 60 seconds.
		shutdownTimeoutCtx, shutdownTimeoutCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer shutdownTimeoutCancel()

		done := make(chan struct{})
		go func() {
			scheduler.Wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("All in-flight jobs completed.")
			logger.Log.Info("All in-flight jobs completed.")
		case <-shutdownTimeoutCtx.Done():
			fmt.Println("Graceful shutdown timeout exceeded. Some jobs might not have completed.")
			logger.Log.Info("Graceful shutdown timeout exceeded. Some jobs might not have completed.")
		}

		fmt.Println("Scheduler exiting.")
		logger.Log.Info("Scheduler exiting.")

		fmt.Println("Clean any dormant or hanging job redis key after scheduler exiting...")
		logger.Log.Info("Clean any dormant or hanging job redis key after scheduler exiting...")
		scheduler.CleanJobRedisKey()
	})

	if err != nil {
		fmt.Printf("container.Invoke Scheduler failed: %v\n", err)
		logger.Log.Errorf("container.Invoke Scheduler failed: %v", err)

		os.Exit(1)
		return
	}

	os.Exit(0)
}

// startHealthServer runs a small HTTP server for health checks
func startHealthServer(ctx context.Context, cfg *config.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", module.APPLICATION_JSON)

		response := map[string]interface{}{
			"status": "ok",
		}

		// Readiness check: Are we in shutdown mode?
		if atomic.LoadInt32(&isShuttingDown) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 Not Ready
			response["status"] = "error"
		} else {
			w.WriteHeader(http.StatusOK) // 200 OK
		}

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			// If marshaling fails, log the error and send a 500 Internal Server Error.
			logger.Log.Errorf("Error marshaling JSON: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 3. Write the JSON bytes to the response writer
		_, err = w.Write(jsonBytes)
		if err != nil {
			logger.Log.Errorf("Error writing JSON response: %v", err)
		}
	})

	server := &http.Server{Addr: cfg.SCHEDULER_HEALTH_SERVER_ADDRESS, Handler: mux}

	// Start health server in a goroutine
	go func() {
		fmt.Printf("Health server starting on port %s...\n", cfg.SCHEDULER_HEALTH_SERVER_ADDRESS)
		logger.Log.Infof("Health server starting on port %s...", cfg.SCHEDULER_HEALTH_SERVER_ADDRESS)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(fmt.Errorf("health server failed: %v", err))
			logger.Log.Fatalf("Health server failed: %v", err)
		}
	}()

	// Listen for main context cancellation to shut down health server
	go func() {
		<-ctx.Done()
		fmt.Println("Main context cancelled, shutting down health server...")
		logger.Log.Info("Main context cancelled, shutting down health server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Println(fmt.Errorf("health server shutdown error: %v", err))
			logger.Log.Errorf("Health server shutdown error: %v", err)
		}
	}()
}
