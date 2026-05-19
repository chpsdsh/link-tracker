package scrapperclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const (
	clientTimeout = 15 * time.Second
)

type UpdatesClient struct {
	Client *http.Client
	Config config.Config
}

func NewUpdatesClient(config config.Config) *UpdatesClient {
	client := &http.Client{Timeout: config.HTTPClientConfig.Timeout}
	return &UpdatesClient{Client: client, Config: config}
}

func (c UpdatesClient) SendLinkUpdate(update pkg.LinkUpdate, _ string) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.Config.BotServerAddr+botPostEndpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error doing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return scrapper.ErrIncorrectRequestParameters
	}
}

func (c UpdatesClient) Close() {
	c.Client.CloseIdleConnections()
}
