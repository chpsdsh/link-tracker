package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

type ScrapperCacheClient struct {
	client *redis.Client
}

func NewScrapperCacheClient(conf config.Config) ScrapperCacheClient {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.ValkeyConfig.Addresses[0],
		Password: conf.ValkeyConfig.Password,
	})
	return ScrapperCacheClient{client: client}
}

func (c ScrapperCacheClient) GetUserLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error) {
	key := userLinksCacheKey(chatID)
	links, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, scrapper.ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get links from cache: %w", err)
	}

	var linksArray []pkg.LinkInfo
	if err = json.Unmarshal([]byte(links), &linksArray); err != nil {
		return nil, fmt.Errorf("failed to unmarshal links: %w", err)
	}
	return linksArray, nil
}

func (c ScrapperCacheClient) SetUserLinks(ctx context.Context, chatID int64, linksArr []pkg.LinkInfo) error {
	key := userLinksCacheKey(chatID)
	links, err := json.Marshal(linksArr)
	if err != nil {
		return fmt.Errorf("failed to marshal links: %w", err)
	}
	if err = c.client.Set(ctx, key, links, 0).Err(); err != nil {
		return fmt.Errorf("failed to set links: %w", err)
	}
	return nil
}

func (c ScrapperCacheClient) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}
	return nil
}

func (c ScrapperCacheClient) DeleteUserLinks(ctx context.Context, chatID int64) error {
	key := userLinksCacheKey(chatID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user links cache: %w", err)
	}
	return nil
}

func userLinksCacheKey(chatID int64) string {
	return fmt.Sprintf("links:list:%d", chatID)
}
