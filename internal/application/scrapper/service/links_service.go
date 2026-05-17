//go:generate mockgen -source links_service.go -destination=../mocks/links_service_mocks.go -package=mocks
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const requestTimeout = 5 * time.Second

type LinkRepository interface {
	AddLink(ctx context.Context, chatID int64, link pkg.LinkInfo) error
	DeleteLink(ctx context.Context, chatID int64, url string) (pkg.LinkInfo, error)
	GetUserLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error)
	GetAllLinks(ctx context.Context, limit int, offset int) ([]pkg.LinkInfo, error)
	UpdateLinksTime(ctx context.Context, newTime time.Time, url string) error
	GetChatIDsByLink(ctx context.Context, link string) ([]int64, error)
	LinkExists(ctx context.Context, url string) (bool, error)
}

type ChatRepository interface {
	ChatExists(ctx context.Context, chatID int64) (bool, error)
	AddChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
}

type Transactor interface {
	Transaction(ctx context.Context, txFunc func(ctx context.Context) error) error
	TransactionWithReturn(ctx context.Context, txFunc func(ctx context.Context) (any, error)) (any, error)
}

type CacheRepository interface {
	GetUserLinks(ctx context.Context, chatID int64, loader func(ctx context.Context, key string) (*[]pkg.LinkInfo, error)) ([]pkg.LinkInfo, error)
	SetUserLinks(ctx context.Context, id int64, links []pkg.LinkInfo) error
	DeleteUserLinks(ctx context.Context, chatID int64) error
}

type LinksService struct {
	LinkRepo   LinkRepository
	ChatsRepo  ChatRepository
	Transactor Transactor
	CacheRepo  CacheRepository
	BaseLogger *slog.Logger
}

func (h LinksService) AddChatID(ctx context.Context, chatID int64) error {
	return h.Transactor.Transaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck// no need to wrap function call
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
		if err != nil {
			return errors.Join(err, scrapper.ErrInternalError)
		}

		if exists {
			return scrapper.ErrChatAlreadyExists
		}

		if err = h.ChatsRepo.AddChat(ctx, chatID); err != nil {
			return errors.Join(err, scrapper.ErrInternalError)
		}
		return nil
	})
}

func (h LinksService) DeleteChat(ctx context.Context, chatID int64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	if err := h.ChatsRepo.DeleteChat(ctx, chatID); err != nil {
		return errors.Join(err, scrapper.ErrInternalError)
	}
	return nil
}

func (h LinksService) GetLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error) {
	cacheLinks, cacheErr := h.CacheRepo.GetUserLinks(ctx, chatID, func(ctx context.Context, key string) (*[]pkg.LinkInfo, error) {
		links, err := h.Transactor.TransactionWithReturn(ctx, func(ctx context.Context) (any, error) {
			ctx, cancel := context.WithTimeout(ctx, requestTimeout)
			defer cancel()
			exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
			if err != nil {
				return nil, errors.Join(err, scrapper.ErrInternalError)
			}
			if !exists {
				return nil, scrapper.ErrChatNotFound
			}

			links, err := h.LinkRepo.GetUserLinks(ctx, chatID)
			if err != nil {
				return nil, errors.Join(err, scrapper.ErrInternalError)
			}
			return links, nil
		})
		if err != nil {
			return nil, errors.Join(err, scrapper.ErrInternalError)
		}

		linksArr, ok := links.([]pkg.LinkInfo)
		if !ok {
			return nil, errors.Join(scrapper.ErrInternalError)
		}
		return &linksArr, nil
	})

	if cacheErr != nil {
		h.BaseLogger.Error("Failed to get links", "error", cacheErr)
		return nil, errors.Join(cacheErr, scrapper.ErrInternalError)
	}

	return cacheLinks, nil
}

func (h LinksService) AddLink(ctx context.Context, chatID int64, linkRequest pkg.AddLinkRequest) error {
	return h.Transactor.Transaction(ctx, func(ctx context.Context) error { //nolint:wrapcheck// no need to wrap function call
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		trackedLink := pkg.LinkInfo{Link: linkRequest.Link, Tags: linkRequest.Tags}
		if err := h.LinkRepo.AddLink(ctx, chatID, trackedLink); err != nil {
			return errors.Join(err, scrapper.ErrInternalError)
		}

		if err := h.CacheRepo.DeleteUserLinks(ctx, chatID); err != nil {
			h.BaseLogger.Error("error deleting links from cache", slog.String("err", err.Error()))
		}

		return nil
	})
}

func (h LinksService) DeleteLink(ctx context.Context, chatID int64, link string) (pkg.LinkInfo, error) {
	linkRes, err := h.Transactor.TransactionWithReturn(ctx, func(ctx context.Context) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
		if err != nil {
			return pkg.LinkInfo{}, errors.Join(err, scrapper.ErrInternalError)
		}
		if !exists {
			return pkg.LinkInfo{}, scrapper.ErrChatNotFound
		}

		exists, err = h.LinkRepo.LinkExists(ctx, link)
		if err != nil {
			return pkg.LinkInfo{}, errors.Join(err, scrapper.ErrInternalError)
		}
		if !exists {
			return pkg.LinkInfo{}, scrapper.ErrLinkNotExists
		}

		linkInfo, err := h.LinkRepo.DeleteLink(ctx, chatID, link)
		if err != nil {
			return pkg.LinkInfo{}, errors.Join(err, scrapper.ErrInternalError)
		}

		if err = h.CacheRepo.DeleteUserLinks(ctx, chatID); err != nil {
			h.BaseLogger.Error("error deleting links from cache", slog.String("err", err.Error()))
		}

		return linkInfo, nil
	})
	if err != nil {
		return pkg.LinkInfo{}, errors.Join(err, scrapper.ErrInternalError)
	}
	linkInfo, ok := linkRes.(pkg.LinkInfo)
	if !ok {
		return pkg.LinkInfo{}, scrapper.ErrInternalError
	}
	return linkInfo, nil
}
