package db

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	// Get Redis configuration from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost" // Default to localhost
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379" // Default Redis port
	}

	redisPassword := os.Getenv("REDIS_PASSWORD") // Can be empty for local development

	redisDB := os.Getenv("REDIS_DB")
	dbNum := 0 // Default database
	if redisDB != "" {
		if db, err := strconv.Atoi(redisDB); err == nil {
			dbNum = db
		}
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisHost + ":" + redisPort,
		Password:     redisPassword,
		DB:           dbNum,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("✅ Redis client ready")
	return rdb, nil
}
