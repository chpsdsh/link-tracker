//go:generate mockgen -source server.go -destination=../mocks/server_mocks.go -package=mocks
package scrapperserver

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	errorReadingRequestBody         = "failed to read request body"
	errorUnmarshallingJSON          = "failed to unmarshal json response"
	errorMarshallingJSON            = "failed to marshal json response"
	errorIncorrectRequestParameters = "incorrect request parameters"
	errorChatNotFound               = "chat not found"
	errorLinkIsAlreadyTracked       = "link is already tracked"
	errorChatAlreadyExists          = "chat already exists"
	errorLinkNotExists              = "link does not exist"
	badRequest                      = "bad_request"
	invalidParameter                = "invalid_parameter"
	missingHeder                    = "missing_header"
	internalError                   = "internal_error"
)

type ScrapperHandler interface {
	AddChatID(chatID int64) error
	DeleteChat(chatID int64) error
	GetLinks(chatID int64) ([]pkg.LinkInfo, error)
	AddLink(chatID int64, linkRequest pkg.AddLinkRequest) error
	DeleteLink(chatID int64, link string) (pkg.LinkInfo, error)
}

type ScrapperServer struct {
	BaseLogger *slog.Logger
	Handler    ScrapperHandler
}

func (s ScrapperServer) DeleteLinks(w http.ResponseWriter, r *http.Request, params DeleteLinksParams) {
	body, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		s.sendAPIErrorResponse(w, errorReadingRequestBody, err, http.StatusBadRequest)
		s.BaseLogger.Error(errorReadingRequestBody, slog.String("error", err.Error()))
		return
	}

	removeRequest := RemoveLinkRequest{}
	if err = json.Unmarshal(body, &removeRequest); err != nil {
		s.sendAPIErrorResponse(w, errorMarshallingJSON, err, http.StatusBadRequest)
		s.BaseLogger.Error(errorMarshallingJSON, slog.String("error", err.Error()))
		return
	}

	link, deleteLinkErr := s.Handler.DeleteLink(params.TgChatId, *removeRequest.Link)

	switch {
	case deleteLinkErr == nil:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		linkResponse := LinkResponse{Tags: &link.Tags, Url: &link.Link}
		data, errMarshall := json.Marshal(linkResponse)
		if errMarshall != nil {
			s.sendAPIErrorResponse(w, errorMarshallingJSON, errMarshall, http.StatusBadRequest)
			s.BaseLogger.Error(errorMarshallingJSON, slog.String("error", errMarshall.Error()))
			return
		}

		if _, errWrite := w.Write(data); errWrite != nil {
			s.BaseLogger.Error("write response failed", slog.String("error", errWrite.Error()))
		}
	case errors.Is(deleteLinkErr, handler.ErrChatNotFound):
		s.sendAPIErrorResponse(w, errorChatNotFound, deleteLinkErr, http.StatusNotFound)
		s.BaseLogger.Error(errorChatNotFound, slog.String("error", deleteLinkErr.Error()))
	case errors.Is(deleteLinkErr, handler.ErrLinkNotExists):
		s.sendAPIErrorResponse(w, errorLinkNotExists, deleteLinkErr, http.StatusNotFound)
		s.BaseLogger.Error(errorLinkNotExists, slog.String("error", deleteLinkErr.Error()))
	}

}

func (s ScrapperServer) GetLinks(w http.ResponseWriter, _ *http.Request, params GetLinksParams) {
	links, err := s.Handler.GetLinks(params.TgChatId)
	switch {
	case err == nil:
		linksResponse := make([]LinkResponse, len(links))

		for i, link := range links {
			linksResponse[i] = LinkResponse{Tags: &link.Tags, Url: &link.Link}
		}
		size := int32(len(linksResponse))
		listLinkResponse := ListLinksResponse{Links: &linksResponse, Size: &size}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		data, errMarshall := json.Marshal(listLinkResponse)
		if errMarshall != nil {
			s.sendAPIErrorResponse(w, errorMarshallingJSON, errMarshall, http.StatusBadRequest)
			s.BaseLogger.Error(errorMarshallingJSON, slog.String("error", errMarshall.Error()))
			return
		}

		if _, errWrite := w.Write(data); errWrite != nil {
			s.BaseLogger.Error("write response failed", slog.String("error", errWrite.Error()))
		}
	case errors.Is(err, handler.ErrChatNotFound):
		s.sendAPIErrorResponse(w, errorChatNotFound, err, http.StatusNotFound)
		s.BaseLogger.Error(errorChatNotFound, slog.String("error", err.Error()))
	}
}

func (s ScrapperServer) PostLinks(w http.ResponseWriter, r *http.Request, params PostLinksParams) {
	body, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		s.sendAPIErrorResponse(w, errorReadingRequestBody, err, http.StatusBadRequest)
		s.BaseLogger.Error(errorReadingRequestBody, slog.String("error", err.Error()))
		return
	}

	linkRequest := AddLinkRequest{}
	if errUnmarshall := json.Unmarshal(body, &linkRequest); errUnmarshall != nil {
		s.sendAPIErrorResponse(w, errorUnmarshallingJSON, errUnmarshall, http.StatusBadRequest)
		s.BaseLogger.Error(errorUnmarshallingJSON, slog.String("error", errUnmarshall.Error()))
		return
	}

	linkErr := s.Handler.AddLink(params.TgChatId, pkg.AddLinkRequest{
		Link: *linkRequest.Link,
		Tags: *linkRequest.Tags,
	})

	switch {
	case linkErr == nil:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		linkResponse := LinkResponse{Filters: linkRequest.Filters, Tags: linkRequest.Tags,
			Id: &params.TgChatId, Url: linkRequest.Link}
		data, errMarshall := json.Marshal(linkResponse)
		if errMarshall != nil {
			s.sendAPIErrorResponse(w, errorMarshallingJSON, errMarshall, http.StatusBadRequest)
			s.BaseLogger.Error(errorMarshallingJSON, slog.String("error", errMarshall.Error()))
			return
		}
		if _, errWrite := w.Write(data); errWrite != nil {
			s.BaseLogger.Error("write response failed", slog.String("error", errWrite.Error()))
		}
	case errors.Is(linkErr, handler.ErrIncorrectRequestParameters):
		s.sendAPIErrorResponse(w, errorIncorrectRequestParameters, linkErr, http.StatusBadRequest)
		s.BaseLogger.Error(errorIncorrectRequestParameters, slog.String("error", linkErr.Error()))
	case errors.Is(linkErr, handler.ErrChatNotFound):
		s.sendAPIErrorResponse(w, errorChatNotFound, linkErr, http.StatusNotFound)
		s.BaseLogger.Error(errorChatNotFound, slog.String("error", linkErr.Error()))
	case errors.Is(linkErr, handler.ErrLinkExists):
		s.sendAPIErrorResponse(w, errorLinkIsAlreadyTracked, linkErr, http.StatusConflict)
		s.BaseLogger.Error(errorLinkIsAlreadyTracked, slog.String("error", linkErr.Error()))
	}
}

func (s ScrapperServer) DeleteTgChatId(w http.ResponseWriter, _ *http.Request, id int64) { //nolint:revive,staticcheck // method name required by oapi-codegen interface
	chatErr := s.Handler.DeleteChat(id)
	switch {
	case chatErr == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(chatErr, handler.ErrChatNotFound):
		s.sendAPIErrorResponse(w, errorChatNotFound, chatErr, http.StatusNotFound)
		s.BaseLogger.Error(errorChatNotFound, slog.String("error", chatErr.Error()))
	}

}

func (s ScrapperServer) PostTgChatId(w http.ResponseWriter, _ *http.Request, id int64) { //nolint:revive,staticcheck // method name required by oapi-codegen interface
	chatErr := s.Handler.AddChatID(id)
	switch {
	case chatErr == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(chatErr, handler.ErrChatAlreadyExists):
		s.sendAPIErrorResponse(w, errorChatAlreadyExists, chatErr, http.StatusConflict)
		s.BaseLogger.Error(errorChatAlreadyExists, slog.String("error", chatErr.Error()))
	}
}

func (s ScrapperServer) sendAPIErrorResponse(w http.ResponseWriter, desc string, err error, status int) {
	code := badRequest
	errString := err.Error()
	errResp := ApiErrorResponse{Code: &code, Description: &desc, ExceptionMessage: &errString}
	data, _ := json.Marshal(errResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, errWrite := w.Write(data); errWrite != nil {
		s.BaseLogger.Error("write response failed", slog.String("error", errWrite.Error()))
	}
}

func JSONErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	var status = http.StatusBadRequest
	code := badRequest

	var invalidParamFormatError *InvalidParamFormatError
	var requiredHeaderError *RequiredHeaderError
	var tooManyValuesForParamError *TooManyValuesForParamError
	switch {

	case errors.As(err, &invalidParamFormatError):
		code = invalidParameter
	case errors.As(err, &requiredHeaderError):
		code = missingHeder
	case errors.As(err, &tooManyValuesForParamError):
		code = invalidParameter
	default:
		status = http.StatusInternalServerError
		code = internalError
	}

	desc := err.Error()

	resp := ApiErrorResponse{
		Code:        &code,
		Description: &desc,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(resp)
}
