package scrapperclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func TestDoGithubRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Accept") != applicationType {
			t.Fatalf("wrong Accept header")
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"updated_at": "2024-03-03T18:58:10Z"
		}`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			GithubToken: "token",
		},
	}

	resp, err := client.DoGithubRequest(server.URL)

	if errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatalf("unexpected error %v", err)
	}
	ti, _ := time.Parse(time.RFC3339, "2024-03-03T18:58:10Z")
	if resp.UpdatedAt != ti {
		t.Fatalf("wrong updated_at %s", resp.UpdatedAt)
	}
}

func TestDoGithubRequestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			GithubToken: "token",
		},
	}

	_, err := client.DoGithubRequest(server.URL)

	if !errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatal("expected json error")
	}
}

func TestDoGithubIssueRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Accept") != applicationType {
			t.Fatalf("wrong Accept header")
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`[
			{
				"title": "issue title",
				"body": "issue body",
				"created_at": "2024-03-03T18:58:10Z",
				"updated_at": "2024-03-04T18:58:10Z",
				"user": {
					"login": "user1"
				}
			}
		]`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			GithubToken: "token",
		},
	}

	resp, err := client.DoGithubIssueRequest(server.URL)

	if errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatalf("unexpected error %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(resp))
	}

	if resp[0].Title != "issue title" {
		t.Fatalf("wrong title %s", resp[0].Title)
	}

	if resp[0].User.Login != "user1" {
		t.Fatalf("wrong user %s", resp[0].User.Login)
	}
}

func TestDoGithubIssueRequest_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			GithubToken: "token",
		},
	}

	_, err := client.DoGithubIssueRequest(server.URL)

	if !errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatalf("expected unmarshalling error, got %v", err)
	}
}

func TestDoGithubPullRequestRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Accept") != applicationType {
			t.Fatalf("wrong Accept header")
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`[
			{
				"title": "pr title",
				"body": "pr body",
				"created_at": "2024-03-03T18:58:10Z",
				"updated_at": "2024-03-05T18:58:10Z",
				"user": {
					"login": "dev1"
				}
			}
		]`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			GithubToken: "token",
		},
	}

	resp, err := client.DoGithubPullRequestRequest(server.URL)

	if errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatalf("unexpected error %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 pull request, got %d", len(resp))
	}

	if resp[0].Title != "pr title" {
		t.Fatalf("wrong title %s", resp[0].Title)
	}

	if resp[0].User.Login != "dev1" {
		t.Fatalf("wrong user %s", resp[0].User.Login)
	}
}

func TestDoStackOverflowRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"items": [
				{
					"last_activity_date": 1710000000
				}
			]
		}`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			StackoverflowToken: "token",
		},
	}

	resp, err := client.DoStackOverflowQuestionRequest(server.URL + "?site=stackoverflow")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}

	if resp.Items[0].LastActivityDate != 1710000000 {
		t.Fatalf("wrong last_activity_date %d", resp.Items[0].LastActivityDate)
	}
}

func TestDoStackOverflowRequestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			StackoverflowToken: "token",
		},
	}

	_, err := client.DoStackOverflowQuestionRequest(server.URL + "?site=stackoverflow")

	if !errors.Is(err, ErrUnmarshallingJSON) {
		t.Fatalf("expected json error, got %v", err)
	}
}

func TestSendLinkUpdateSuccess(t *testing.T) {

	update := pkg.LinkUpdate{
		Description: "link updated",
		TgChatIDs:   []int64{1, 2},
		URL:         "https://github.com/golang/go",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.Header.Get(contentTypeKey) != typeApplicationJSON {
			t.Fatalf("expected content-type %s", typeApplicationJSON)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var received pkg.LinkUpdate
		if err = json.Unmarshal(body, &received); err != nil {
			t.Fatal(err)
		}

		if received.URL != update.URL {
			t.Fatalf("expected url %s, got %s", update.URL, received.URL)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := UpdatesClient{
		Client: server.Client(),
		Config: config.Config{BotServerAddr: server.URL},
	}

	err := client.SendLinkUpdate(update, "")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestSendLinkUpdateBadStatus(t *testing.T) {
	update := pkg.LinkUpdate{
		Description: "link updated",
		TgChatIDs:   []int64{1},
		URL:         "https://github.com/golang/go",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := UpdatesClient{
		Client: server.Client(),
		Config: config.Config{BotServerAddr: server.URL},
	}

	err := client.SendLinkUpdate(update, "")

	if !errors.Is(err, scrapper.ErrIncorrectRequestParameters) {
		t.Fatalf("expected ErrIncorrectRequestParameters, got %v", err)
	}
}

func TestDoStackOverflowAnswersRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"items": [
				{
					"last_activity_date": 1710000000,
					"creation_date": 1700000000,
					"body": "test answer",
					"owner": {
						"display_name": "user1"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			StackoverflowToken: "token",
		},
	}

	resp, err := client.DoStackOverflowAnswersRequest(server.URL + "?site=stackoverflow")
	require.NoError(t, err)

	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1710000000), resp.Items[0].LastActivityDate)
	assert.Equal(t, "test answer", resp.Items[0].Body)
	assert.Equal(t, "user1", resp.Items[0].Owner.DisplayName)
}

func TestDoStackOverflowAnswersRequest_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			StackoverflowToken: "token",
		},
	}

	_, err := client.DoStackOverflowAnswersRequest(server.URL + "?site=stackoverflow")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnmarshallingJSON)
}

func TestDoStackOverflowCommentsRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"items": [
				{
					"last_activity_date": 1720000000,
					"creation_date": 1710000000,
					"body": "test comment",
					"owner": {
						"display_name": "comment_user"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := Client{
		Client: server.Client(),
		Config: config.Config{
			StackoverflowToken: "token",
		},
	}

	resp, err := client.DoStackOverflowCommentsRequest(server.URL + "?site=stackoverflow")
	require.NoError(t, err)

	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1710000000), resp.Items[0].CreationDate)
	assert.Equal(t, "test comment", resp.Items[0].Body)
	assert.Equal(t, "comment_user", resp.Items[0].Owner.DisplayName)
}
