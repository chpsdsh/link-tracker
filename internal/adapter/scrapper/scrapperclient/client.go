package scrapperclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/scheduler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

var JsonUnmarshallingError = errors.New("json unmarshalling error")

const (
	botPostEndpoint     = "/updates"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
	applicationType     = "application/vnd.github.v3+json"
	version             = "2022-11-28"
	stackOverflowKey    = "&key="
)

type Client struct {
	Client *http.Client
	Config config.Config
}

func (c Client) DoGithubRequest(url string) (scrapper.GitHubUpdate, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return scrapper.GitHubUpdate{}, err
	}

	req.Header.Add("Accept", applicationType)
	req.Header.Add("Authorization", c.Config.GithubToken)
	req.Header.Add("X-GitHub-Api-Version", version)

	resp, err := c.Client.Do(req)
	if err != nil {
		return scrapper.GitHubUpdate{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return scrapper.GitHubUpdate{}, err
	}
	gitUpdate := scrapper.GitHubUpdate{}
	if err := json.Unmarshal(data, &gitUpdate); err != nil {
		return scrapper.GitHubUpdate{}, errors.Join(err, JsonUnmarshallingError)
	}
	return gitUpdate, nil
}

func (c Client) DoStackOverflowRequest(url string) (scrapper.StackOverflowUpdate, error) {
	req, err := http.NewRequest(http.MethodGet, url+stackOverflowKey+c.Config.StackoverflowToken, nil)
	if err != nil {
		return scrapper.StackOverflowUpdate{}, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return scrapper.StackOverflowUpdate{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return scrapper.StackOverflowUpdate{}, err
	}

	stackOverflowUpdate := scrapper.StackOverflowUpdate{}
	if err := json.Unmarshal(data, &stackOverflowUpdate); err != nil {
		return scrapper.StackOverflowUpdate{}, errors.Join(err, JsonUnmarshallingError)
	}

	return stackOverflowUpdate, nil
}

func (c Client) SendLinkUpdate(update shared.LinkUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.Config.BotServerAddr+botPostEndpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return scheduler.IncorrectRequestParametersError
	}
}
