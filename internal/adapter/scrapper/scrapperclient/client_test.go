package scrapperclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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

	if resp.UpdatedAt != "2024-03-03T18:58:10Z" {
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

	resp, err := client.DoStackOverflowRequest(server.URL + "?site=stackoverflow")
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

	_, err := client.DoStackOverflowRequest(server.URL + "?site=stackoverflow")

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

	client := Client{
		Client: server.Client(),
		Config: config.Config{BotServerAddr: server.URL},
	}

	err := client.SendLinkUpdate(update)

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

	client := Client{
		Client: server.Client(),
		Config: config.Config{BotServerAddr: server.URL},
	}

	err := client.SendLinkUpdate(update)

	if !errors.Is(err, scrapper.ErrIncorrectRequestParameters) {
		t.Fatalf("expected ErrIncorrectRequestParameters, got %v", err)
	}
}
