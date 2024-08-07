// redis.go
package main

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Addr of redis server
	})
}

// GetBalanceFromCache get balance from cache Redis
func GetBalanceFromCache(accountID string) (float64, error) {
	cachedBalance, err := rdb.Get(ctx, accountID).Result()
	if errors.Is(err, redis.Nil) {
		return 0, err
	} else if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(cachedBalance, 64)
}

// SetBalanceToCache set balance to cache Redis
func SetBalanceToCache(accountID string, balance float64) error {
	err := rdb.Set(ctx, accountID, balance, 5*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to set balance in Redis for account %s: %v\n", accountID, err)
		return err
	}
	return nil
}
