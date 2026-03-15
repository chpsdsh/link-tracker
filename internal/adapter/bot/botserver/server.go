//go:generate mockgen -source server.go -destination=../mocks/server_mocks.go -package=mocks
package botserver

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	errorUnmarshallingJSON  = "failed to unmarshal json response"
	errorReadingRequestBody = "failed to read request body"
	badRequestCode          = "BAD_REQUEST"
)

type TelegramBotHandler interface {
	HandleLinkUpdate(linkUpdate pkg.LinkUpdate)
}

type UpdatesServer struct {
	BaseLogger *slog.Logger
	Handler    TelegramBotHandler
}

func (u *UpdatesServer) PostUpdates(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		sendAPIErrorResponse(w, errorReadingRequestBody, err)
		u.BaseLogger.Error(errorReadingRequestBody, slog.String("error", err.Error()))
		return
	}

	linkUpdate := LinkUpdate{}
	if err = json.Unmarshal(body, &linkUpdate); err != nil {
		sendAPIErrorResponse(w, errorUnmarshallingJSON, err)
		u.BaseLogger.Error(errorUnmarshallingJSON, slog.String("error", err.Error()))
		return
	}

	if linkUpdate.Description == nil || linkUpdate.TgChatIds == nil || linkUpdate.Url == nil {
		sendAPIErrorResponse(w, "missing required fields", errors.New("description, tgChatIds and url are required"))
		return
	}

	u.BaseLogger.Info("response ", slog.Any("update", linkUpdate))
	u.Handler.HandleLinkUpdate(pkg.LinkUpdate{Description: *linkUpdate.Description,
		TgChatIDs: *linkUpdate.TgChatIds, URL: *linkUpdate.Url})

	w.WriteHeader(http.StatusOK)
}

func sendAPIErrorResponse(w http.ResponseWriter, desc string, err error) {
	code := badRequestCode
	errString := err.Error()
	errResp := ApiErrorResponse{Code: &code, Description: &desc, ExceptionMessage: &errString}
	data, _ := json.Marshal(errResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(data)
}
