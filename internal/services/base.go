package services

import (
	"context"
	"fmt"
	"go-api-boilerplate/internal/constant"
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/db"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"go-api-boilerplate/module/redis"
	"strings"
)

type Service struct {
	Config   *config.Config
	DB       *db.DBConnection
	Redis    *redis.RedisConnection
	Elastic  elastic.ElasticConnections
	Response *response.Response
}

func NewService(cfg *config.Config, dbc *db.DBConnection, rdc *redis.RedisConnection, es elastic.ElasticConnections, resp *response.Response) *Service {
	return &Service{
		Config:   cfg,
		DB:       dbc,
		Redis:    rdc,
		Elastic:  es,
		Response: resp,
	}
}

func (s *Service) ClearCacheUserInfo(ctx context.Context, userID int) {
	key := fmt.Sprintf(constant.USER_INFO, userID)

	s.Redis.DelCache(ctx, key)
}

func (s *Service) ClearCacheUser(ctx context.Context, userID int) {
	keys := []string{
		fmt.Sprintf(constant.USER_JWT, userID),
		fmt.Sprintf(constant.USER_INFO, userID),
	}

	s.Redis.DelCache(ctx, keys...)
}

func (s *Service) NXUnlock(ctx context.Context, lockRedisKey string) error {
	if err := s.Redis.DelCache(ctx, lockRedisKey); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"redis_key": lockRedisKey,
			"action":    "DELETE",
		})

		return err
	}

	return nil
}

func (s *Service) MultipleNXUnlock(ctx context.Context, lockRedisKeys []string) error {
	if err := s.Redis.DelCache(ctx, lockRedisKeys...); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"redis_key": strings.Join(lockRedisKeys, ", "),
			"action":    "DELETE",
		})

		return err
	}

	return nil
}
