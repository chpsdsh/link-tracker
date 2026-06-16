package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrSerializingLinks   = errors.New("error serializing links")
	ErrDeserializingLinks = errors.New("error deserializing links")
	ErrNilLinks           = errors.New("get user links with cache-aside: nil links")
)

type ScrapperCacheClient struct {
	baseClient  valkey.Client
	asideClient valkeyaside.TypedCacheAsideClient[[]pkg.LinkInfo]
	ttl         time.Duration
}

func NewScrapperCacheClient(conf config.ScrapperConfig) (ScrapperCacheClient, error) {
	baseClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: conf.ValkeyConfig.Addresses,
		Password:    conf.ValkeyConfig.Password,
	})
	if err != nil {
		return ScrapperCacheClient{}, fmt.Errorf("create valkey client: %w", err)
	}

	client, err := valkeyaside.NewClient(valkeyaside.ClientOption{
		ClientOption: valkey.ClientOption{
			InitAddress: []string{conf.ValkeyConfig.Addresses[0]},
			Password:    conf.ValkeyConfig.Password,
		},
	})
	if err != nil {
		return ScrapperCacheClient{}, fmt.Errorf("create valkey client: %w", err)
	}

	serializer := func(links *[]pkg.LinkInfo) (string, error) {
		data, errMarshal := json.Marshal(*links)
		if errMarshal != nil {
			return "", errors.Join(errMarshal, ErrSerializingLinks)
		}
		return string(data), nil
	}

	deserializer := func(data string) (*[]pkg.LinkInfo, error) {
		var links []pkg.LinkInfo
		if errUnmarshal := json.Unmarshal([]byte(data), &links); errUnmarshal != nil {
			return nil, errors.Join(errUnmarshal, ErrDeserializingLinks)
		}
		return &links, nil
	}

	asideClient := valkeyaside.NewTypedCacheAsideClient(
		client,
		serializer,
		deserializer,
	)

	return ScrapperCacheClient{baseClient: baseClient, asideClient: asideClient, ttl: conf.ValkeyConfig.ValkeyTTL}, nil
}

func (c ScrapperCacheClient) GetUserLinks(ctx context.Context, chatID int64, loader func(ctx context.Context, key string) (*[]pkg.LinkInfo, error)) ([]pkg.LinkInfo, error) {
	key := userLinksCacheKey(chatID)
	links, err := c.asideClient.Get(ctx, c.ttl, key, loader)
	if err != nil {
		return nil, fmt.Errorf("failed to get links from cache: %w", err)
	}

	if links == nil {
		return nil, ErrNilLinks
	}
	return *links, nil
}

func (c ScrapperCacheClient) SetUserLinks(ctx context.Context, chatID int64, linksArr []pkg.LinkInfo) error {
	key := userLinksCacheKey(chatID)
	links, err := json.Marshal(linksArr)
	if err != nil {
		return fmt.Errorf("failed to marshal links: %w", err)
	}

	if err = c.baseClient.Do(
		ctx,
		c.baseClient.B().
			Set().
			Key(key).
			Value(string(links)).
			Build(),
	).Error(); err != nil {
		return fmt.Errorf("failed to set links: %w", err)
	}
	return nil
}

func (c ScrapperCacheClient) Close() error {
	c.baseClient.Close()
	c.asideClient.Client().Close()
	return nil
}

func (c ScrapperCacheClient) DeleteUserLinks(ctx context.Context, chatID int64) error {
	key := userLinksCacheKey(chatID)
	if err := c.baseClient.Do(ctx,
		c.baseClient.
			B().
			Del().
			Key(key).
			Build(),
	).Error(); err != nil {
		return fmt.Errorf("failed to delete user links cache: %w", err)
	}
	return nil
}

func userLinksCacheKey(chatID int64) string {
	return fmt.Sprintf("links:list:%d", chatID)
}
