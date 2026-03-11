package scrapperclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/scheduler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

const (
	botPostAddress      = "http://localhost:8080/updates"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
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

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", c.Config.GithubToken)
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

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
		return scrapper.GitHubUpdate{}, err
	}
	return gitUpdate, nil
}

func (c Client) SendLinkUpdate(update shared.LinkUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, botPostAddress, bytes.NewReader(data))
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
	case http.StatusBadRequest:
		return scheduler.IncorrectRequestParametersError
	}
	return nil
}
