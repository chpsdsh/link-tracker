package httpsender

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

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
