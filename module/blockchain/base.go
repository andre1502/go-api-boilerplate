package blockchain

import (
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/redis"
	"strings"
)

type Blockchain struct {
	config *config.Config
	redis  *redis.RedisConnection
}

func NewBlockchain(
	cfg *config.Config,
	rds *redis.RedisConnection,
) *Blockchain {
	return &Blockchain{
		config: cfg,
		redis:  rds,
	}
}

func (b *Blockchain) RemoveHexPrefix(str string) string {
	if strings.HasPrefix(str, "0x") {
		return str[2:]
	}

	return str
}
