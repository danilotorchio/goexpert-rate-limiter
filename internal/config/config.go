package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis    RedisConfig
	Server   ServerConfig
	RateLimit RateLimitConfig
	Tokens   map[string]int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
}

type RateLimitConfig struct {
	DefaultIPLimit    int
	DefaultTokenLimit int
	BlockDuration     time.Duration
}

func Load() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	defaultIPLimit, _ := strconv.Atoi(getEnv("DEFAULT_IP_LIMIT", "10"))
	defaultTokenLimit, _ := strconv.Atoi(getEnv("DEFAULT_TOKEN_LIMIT", "100"))
	blockDurationSeconds, _ := strconv.Atoi(getEnv("BLOCK_DURATION_SECONDS", "300"))

	cfg := &Config{
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RateLimit: RateLimitConfig{
			DefaultIPLimit:    defaultIPLimit,
			DefaultTokenLimit: defaultTokenLimit,
			BlockDuration:     time.Duration(blockDurationSeconds) * time.Second,
		},
		Tokens: loadTokenConfig(),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadTokenConfig() map[string]int {
	tokens := make(map[string]int)
	
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		
		key := pair[0]
		value := pair[1]
		
		if strings.HasPrefix(key, "TOKEN_") && strings.HasSuffix(key, "_LIMIT") {
			tokenName := strings.TrimSuffix(strings.TrimPrefix(key, "TOKEN_"), "_LIMIT")
			if limit, err := strconv.Atoi(value); err == nil {
				tokens[tokenName] = limit
			}
		}
	}
	
	return tokens
} 