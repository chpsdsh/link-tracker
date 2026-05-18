package scrapperclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

var ErrUnmarshallingJSON = errors.New("json unmarshalling error")

const (
	botPostEndpoint     = "/updates"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
	applicationType     = "application/vnd.github.v3+json"
	version             = "2026-03-10"
	stackOverflowKey    = "&key="
)

type Client struct {
	Client *http.Client
	Config config.Config
}

func (c Client) DoGithubRequest(url string) (scrapper.GitHubRepositoryResponse, error) {
	resp, err := c.doGithubAPIRequest(url)
	if err != nil {
		return scrapper.GitHubRepositoryResponse{}, fmt.Errorf("error requesting github API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return scrapper.GitHubRepositoryResponse{}, fmt.Errorf("error reading response body: %w", err)
	}
	gitUpdate := scrapper.GitHubRepositoryResponse{}
	if err = json.Unmarshal(data, &gitUpdate); err != nil {
		return scrapper.GitHubRepositoryResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}
	return gitUpdate, nil
}

func (c Client) DoGithubIssueRequest(url string) ([]scrapper.GithubIssue, error) {
	resp, err := c.doGithubAPIRequest(url)
	if err != nil {
		return nil, fmt.Errorf("error requesting github API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var issueResponse []scrapper.GithubIssue
	if err = json.Unmarshal(data, &issueResponse); err != nil {
		return nil, errors.Join(err, ErrUnmarshallingJSON)
	}
	return issueResponse, nil
}

func (c Client) DoGithubPullRequestRequest(url string) ([]scrapper.GithubPullRequest, error) {
	resp, err := c.doGithubAPIRequest(url)
	if err != nil {
		return nil, fmt.Errorf("error requesting github API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	var pullRequestsResponse []scrapper.GithubPullRequest
	if err = json.Unmarshal(data, &pullRequestsResponse); err != nil {
		return nil, errors.Join(err, ErrUnmarshallingJSON)
	}
	return pullRequestsResponse, nil
}

func (c Client) DoStackOverflowQuestionRequest(url string) (scrapper.StackOverflowQuestionResponse, error) {
	resp, err := c.doStackoverflowAPIRequest(url)
	if err != nil {
		return scrapper.StackOverflowQuestionResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return scrapper.StackOverflowQuestionResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	stackOverflowUpdate := scrapper.StackOverflowQuestionResponse{}
	if err = json.Unmarshal(data, &stackOverflowUpdate); err != nil {
		return scrapper.StackOverflowQuestionResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowUpdate, nil
}

func (c Client) DoStackOverflowAnswersRequest(url string) (scrapper.StackOverflowAnswersResponse, error) {
	resp, err := c.doStackoverflowAPIRequest(url)
	if err != nil {
		return scrapper.StackOverflowAnswersResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return scrapper.StackOverflowAnswersResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	stackOverflowAnswers := scrapper.StackOverflowAnswersResponse{}
	if err = json.Unmarshal(data, &stackOverflowAnswers); err != nil {
		return scrapper.StackOverflowAnswersResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowAnswers, nil
}

func (c Client) DoStackOverflowCommentsRequest(url string) (scrapper.StackOverflowCommentsResponse, error) {
	resp, err := c.doStackoverflowAPIRequest(url)
	if err != nil {
		return scrapper.StackOverflowCommentsResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return scrapper.StackOverflowCommentsResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	stackOverflowComments := scrapper.StackOverflowCommentsResponse{}
	if err = json.Unmarshal(data, &stackOverflowComments); err != nil {
		return scrapper.StackOverflowCommentsResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowComments, nil
}

func (c Client) doGithubAPIRequest(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Accept", applicationType)
	req.Header.Add("Authorization", c.Config.GithubToken)
	req.Header.Add("X-GitHub-Api-Version", version)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	return resp, nil
}

func (c Client) doStackoverflowAPIRequest(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url+stackOverflowKey+c.Config.StackoverflowToken, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	return resp, nil
}
