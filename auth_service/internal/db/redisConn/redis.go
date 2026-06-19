package redisConn

import (
	"context"
	"encoding/json"
	"time"

	"auth_service/pkg/log"
	"auth_service/internal/config"
	"auth_service/internal/models"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

const (
	CacheTTLTasks = 5 * time.Minute
	CacheTTLProfile = 15 * time.Minute
)

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		Password: config.RedisPassword,
		DB: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		log.Logger.Fatal().Err(err).Msg("Redis connection error")
	}

	log.Logger.Info().Msg("Connected to redis")
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
		log.Logger.Info().Msg("Redis connection closed")
	}
}

func SetCache(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RedisClient.Set(ctx, key, data, ttl).Err()
}

func GetCache(key string, dest interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func DeleteCache(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	return RedisClient.Del(ctx, key).Err()
}

func TasksListKey(userID string) string {
	return "tasks:list:" + userID
}

func UserProfileKey(userID string) string {
	return "user:profile:" + userID
}

func CacheUserProfile(userID string, user models.User) error {
	return SetCache(UserProfileKey(userID), user, CacheTTLProfile)
}

func GetCachedUserProfile(userID string) (*models.User, error) {
	var user models.User
	err := GetCache(UserProfileKey(userID), &user)
	return &user, err
}

func InvalidateUserProfile(userID string) error {
	return DeleteCache(UserProfileKey(userID))
}
