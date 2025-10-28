package cmd

import (
	"fmt"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/internal/validation"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/db"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"go-api-boilerplate/module/pagination"
	"go-api-boilerplate/module/redis"
	"strings"
	"time"

	"go.uber.org/dig"
)

func MainInit(repoName, branchName, commitHash, buildDate, version string, isShuttingDown *int32) *dig.Container {
	fmt.Println("start main")

	exception.NewException()

	container := dig.New()

	if module.IsEmptyString(buildDate) {
		buildDate = time.Now().Format(module.DATETIME_ELASTIC_FORMAT)
	}

	cfg, err := config.LoadConfig(repoName, branchName, commitHash, buildDate, version, isShuttingDown)
	if err != nil {
		fmt.Printf("Error load config: %v\n", err)
		panic("Error load config")
	}

	// 註冊配置
	container.Provide(func() *config.Config {
		return cfg
	})

	logger.NewLogger(strings.ToLower(cfg.APP_NAME))

	es, err := elastic.NewElasticConnection(cfg)
	if err != nil {
		fmt.Println(err)
		logger.Log.Fatal(err)
	}

	// Elastic
	container.Provide(func() elastic.ElasticConnections {
		return es
	})

	// DB
	container.Provide(db.NewDB)

	// Redis
	container.Provide(redis.NewRedisConnections)

	// Validation
	container.Provide(validation.NewValidation)

	// Pagination
	container.Provide(pagination.NewPagination)

	// Response
	container.Provide(response.NewResponse)

	InitInternalContainer(container)
	InitPlatformContainer(container)
	InitBackOfficeContainer(container)

	return container
}
