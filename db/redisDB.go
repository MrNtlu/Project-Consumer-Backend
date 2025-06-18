package db

import (
	"context"
	"crypto/tls"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	// Try to get REDIS_URL first (Heroku/Upstash style)
	redisURL := os.Getenv("REDIS_URL")

	var rdb *redis.Client

	if redisURL != "" {
		// Parse Redis URL (supports redis:// and rediss:// schemes)
		parsedURL, err := url.Parse(redisURL)
		if err != nil {
			return nil, err
		}

		// Extract password
		password := ""
		if parsedURL.User != nil {
			password, _ = parsedURL.User.Password()
		}

		// Extract database number from path
		dbNum := 0
		if len(parsedURL.Path) > 1 {
			if db, err := strconv.Atoi(parsedURL.Path[1:]); err == nil {
				dbNum = db
			}
		}

		// Configure TLS for rediss:// scheme
		var tlsConfig *tls.Config
		if parsedURL.Scheme == "rediss" {
			tlsConfig = &tls.Config{
				ServerName: parsedURL.Hostname(),
			}
		}

		rdb = redis.NewClient(&redis.Options{
			Addr:         parsedURL.Host,
			Password:     password,
			DB:           dbNum,
			TLSConfig:    tlsConfig,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})
	} else {
		// Fallback to individual environment variables
		redisHost := os.Getenv("REDIS_HOST")
		if redisHost == "" {
			redisHost = "localhost"
		}

		redisPort := os.Getenv("REDIS_PORT")
		if redisPort == "" {
			redisPort = "6379"
		}

		redisPassword := os.Getenv("REDIS_PASSWORD")

		redisDB := os.Getenv("REDIS_DB")
		dbNum := 0
		if redisDB != "" {
			if db, err := strconv.Atoi(redisDB); err == nil {
				dbNum = db
			}
		}

		rdb = redis.NewClient(&redis.Options{
			Addr:         redisHost + ":" + redisPort,
			Password:     redisPassword,
			DB:           dbNum,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("✅ Redis client ready")
	return rdb, nil
}
