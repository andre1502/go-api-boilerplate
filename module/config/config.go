package config

import (
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	START_TIME                      time.Time
	IS_SHUTTING_DOWN                *int32
	HOST_NAME                       string
	HOST_IP                         string
	POD_ID                          string
	POD_NAME                        string
	POD_IP                          string
	REPO_NAME                       string
	BRANCH_NAME                     string
	COMMIT_HASH                     string
	BUILD_DATE                      string
	VERSION                         string
	APP_NAME                        string
	SERVER_URL                      string
	SERVER_ADDRESS                  string
	SCHEDULER_HEALTH_SERVER_ADDRESS string
	DB                              *DbConfig
	REDIS                           *RedisConfig
	JWT_SECRET_KEY                  string
	PLATFORM_JWT_EXPIRE_MINUTES     string
	BACKOFFICE_JWT_EXPIRE_MINUTES   string
	ACCOUNT_ID_LENGTH               string
	ELASTIC_URL                     string
	ELASTIC_USERNAME                string
	ELASTIC_PASSWORD                string
	KAIA_ENDPOINT                   string
	KAIA_SENDER_PRIVATE_KEY         string
	KAIA_TEST_MODE                  string
	SCHEDULE_REDIS_PREFIX           string
	MAX_FILE_SIZE_MB                string
	MAX_FILE_COUNT                  string
	GCP_CLOUD_STORAGE_BUCKET_NAME   string
}

type DbConfig struct {
	DB_URL                       string
	DB_MAX_OPEN_CONNS            string
	DB_MAX_IDLE_CONNS            string
	DB_CONN_MAX_IDLE_TIME_SECOND string
	DB_CONN_MAX_LIFE_TIME_SECOND string
}

type RedisConfig struct {
	REDIS_NAME     string
	REDIS_ADDR     string
	REDIS_PASSWORD string
	REDIS_DB       string
	REDIS_PREFIX   string
}

// 读取配置
func LoadConfig(repoName, branchName, commitHash, buildDate, version string, isShuttingDown *int32) (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		START_TIME:                      time.Now(),
		IS_SHUTTING_DOWN:                isShuttingDown,
		HOST_NAME:                       hostName,
		REPO_NAME:                       repoName,
		POD_ID:                          os.Getenv("POD_ID"),
		POD_NAME:                        os.Getenv("POD_NAME"),
		POD_IP:                          os.Getenv("POD_IP"),
		BRANCH_NAME:                     branchName,
		COMMIT_HASH:                     commitHash,
		BUILD_DATE:                      buildDate,
		VERSION:                         version,
		APP_NAME:                        getEnv("APP_NAME", "APP"),
		SERVER_URL:                      getEnv("SERVER_URL", ""),
		SERVER_ADDRESS:                  getEnv("SERVER_ADDRESS", "8089"),
		SCHEDULER_HEALTH_SERVER_ADDRESS: getEnv("SCHEDULER_HEALTH_SERVER_ADDRESS", "8080"),
		JWT_SECRET_KEY:                  getEnv("JWT_SECRET_KEY", ""),
		PLATFORM_JWT_EXPIRE_MINUTES:     getEnv("PLATFORM_JWT_EXPIRE_MINUTES", "4320"),
		BACKOFFICE_JWT_EXPIRE_MINUTES:   getEnv("BACKOFFICE_JWT_EXPIRE_MINUTES", "2880"),
		ACCOUNT_ID_LENGTH:               getEnv("ACCOUNT_ID_LENGTH", "8"),
		ELASTIC_URL:                     getEnv("ELASTIC_URL", ""),
		ELASTIC_USERNAME:                getEnv("ELASTIC_USERNAME", ""),
		ELASTIC_PASSWORD:                getEnv("ELASTIC_PASSWORD", ""),
		KAIA_ENDPOINT:                   getEnv("KAIA_ENDPOINT", ""),
		KAIA_SENDER_PRIVATE_KEY:         getEnv("KAIA_SENDER_PRIVATE_KEY", ""),
		KAIA_TEST_MODE:                  getEnv("KAIA_TEST_MODE", ""),
		SCHEDULE_REDIS_PREFIX:           getEnv("SCHEDULE_REDIS_PREFIX", "Schedule:"),
		MAX_FILE_SIZE_MB:                getEnv("MAX_FILE_SIZE_MB", "5"),
		MAX_FILE_COUNT:                  getEnv("MAX_FILE_COUNT", "10"),
		GCP_CLOUD_STORAGE_BUCKET_NAME:   getEnv("GCP_CLOUD_STORAGE_BUCKET_NAME", ""),
	}

	hostIP, err := cfg.GetOutboundIP()
	if err != nil {
		return nil, err
	}

	cfg.HOST_IP = hostIP.String()
	cfg.loadDbConfig()
	cfg.loadRedisConfig()

	return cfg, nil
}

func (cfg *Config) loadDbConfig() {
	cfg.DB = &DbConfig{
		DB_URL:                       getEnv("DATABASE_URL", ""),
		DB_MAX_OPEN_CONNS:            getEnv("DATABASE_MAX_OPEN_CONNS", "100"),
		DB_MAX_IDLE_CONNS:            getEnv("DATABASE_MAX_IDLE_CONNS", "10"),
		DB_CONN_MAX_IDLE_TIME_SECOND: getEnv("DATABASE_CONN_MAX_IDLE_TIME_SECOND", "600"),
		DB_CONN_MAX_LIFE_TIME_SECOND: getEnv("DATABASE_CONN_MAX_LIFE_TIME_SECOND", "30"),
	}
}

func (cfg *Config) loadRedisConfig() {
	cfg.REDIS = &RedisConfig{
		REDIS_ADDR:     getEnv("REDIS_ADDR", ""),
		REDIS_PASSWORD: getEnv("REDIS_PASSWORD", ""),
		REDIS_DB:       getEnv("REDIS_DB", ""),
		REDIS_PREFIX:   getEnv("REDIS_PREFIX", ""),
	}
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// Get preferred outbound ip of this machine
func (cfg *Config) GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
