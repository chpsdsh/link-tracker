package botserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

const (
	errorUnmarshallingJson  = "failed to unmarshal json response"
	errorReadingRequestBody = "failed to read request body"
)

type UpdatesServer struct {
	BaseLogger *slog.Logger
	Handler    handler.TelegramHandler
}

func (u *UpdatesServer) PostUpdates(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		sendApiErrorResponse(w, errorReadingRequestBody, err)
		u.BaseLogger.Error(errorReadingRequestBody, slog.String("error", err.Error()))
		return
	}

	linkUpdate := shared.LinkUpdate{}
	if err = json.Unmarshal(body, &linkUpdate); err != nil {
		sendApiErrorResponse(w, errorUnmarshallingJson, err)
		u.BaseLogger.Error(errorUnmarshallingJson, slog.String("error", err.Error()))
		return
	}

	u.BaseLogger.Info("response ", slog.Any("update", linkUpdate))
	u.Handler.HandleLinkUpdate(linkUpdate)

	w.WriteHeader(http.StatusOK)
}

func sendApiErrorResponse(w http.ResponseWriter, desc string, err error) {
	code := "BAD_REQUEST"
	errString := err.Error()
	errResp := shared.ApiErrorResponse{Code: &code, Description: &desc, ExceptionMessage: &errString}
	data, _ := json.Marshal(errResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(data)
}
