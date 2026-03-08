package scrapperserver

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const (
	errorReadingRequestBody         = "failed to read request body"
	errorUnmarshallingJson          = "failed to unmarshal json response"
	errorMarshallingJson            = "failed to marshal json response"
	errorIncorrectRequestParameters = "incorrect request parameters"
	errorChatNotFound               = "chat not found"
	errorLinkIsAlreadyTracked       = "link not found"
	badRequestCode                  = "BAD_REQUEST"
)

type ScrapperServer struct {
	BaseLogger *slog.Logger
	Handler    handler.LinksHandler
}

func (s ScrapperServer) DeleteLinks(w http.ResponseWriter, r *http.Request, params DeleteLinksParams) {

}

func (s ScrapperServer) GetLinks(w http.ResponseWriter, r *http.Request, params GetLinksParams) {

}

func (s ScrapperServer) PostLinks(w http.ResponseWriter, r *http.Request, params PostLinksParams) {
	body, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		sendApiErrorResponse(w, errorReadingRequestBody, err)
		s.BaseLogger.Error(errorReadingRequestBody, slog.String("error", err.Error()))
		return
	}

	linkRequest := scrapper.AddLinkRequest{}
	if err = json.Unmarshal(body, &linkRequest); err != nil {
		sendApiErrorResponse(w, errorUnmarshallingJson, err)
		s.BaseLogger.Error(errorUnmarshallingJson, slog.String("error", err.Error()))
		return
	}

	code := s.Handler.AddLink(params.TgChatId, linkRequest)
	switch code {
	case http.StatusOK:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		linkResponse := scrapper.LinkResponse{Filters: linkRequest.Filters, Tags: linkRequest.Tags,
			Id: &params.TgChatId, Url: linkRequest.Link}
		data, err := json.Marshal(linkResponse)
		if err != nil {
			sendApiErrorResponse(w, errorMarshallingJson, err)
			s.BaseLogger.Error(errorMarshallingJson, slog.String("error", err.Error()))
			return
		}
		if _, err := w.Write(data); err != nil {
			s.BaseLogger.Error("write response failed", "err", err)
		}
	case http.StatusBadRequest:
		w.WriteHeader(http.StatusNotFound)
		err := errors.New("invalid parameters")
		sendApiErrorResponse(w, errorIncorrectRequestParameters, err)
		s.BaseLogger.Error(errorIncorrectRequestParameters, slog.String("error", err.Error()))
	case http.StatusNotFound:
		w.WriteHeader(http.StatusNotFound)
		err := errors.New("chat not found")
		sendApiErrorResponse(w, errorChatNotFound, err)
		s.BaseLogger.Error(errorChatNotFound, slog.String("error", err.Error()))
	case http.StatusConflict:
		w.WriteHeader(http.StatusConflict)
		err := errors.New("link is already tracked")
		sendApiErrorResponse(w, errorLinkIsAlreadyTracked, err)
		s.BaseLogger.Error(errorLinkIsAlreadyTracked, slog.String("error", err.Error()))
	}

}

func (s ScrapperServer) DeleteTgChatId(w http.ResponseWriter, r *http.Request, id int64) {

}

func (s ScrapperServer) PostTgChatId(w http.ResponseWriter, r *http.Request, id int64) {

}

func sendApiErrorResponse(w http.ResponseWriter, desc string, err error) {
	code := badRequestCode
	errString := err.Error()
	errResp := ApiErrorResponse{Code: &code, Description: &desc, ExceptionMessage: &errString}
	data, _ := json.Marshal(errResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(data)
}
