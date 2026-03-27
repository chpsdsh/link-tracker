//go:generate mockgen -source links_handler.go -destination=../mocks/links_handler_mocks.go -package=mocks
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrChatNotFound               = errors.New("chat not found")
	ErrLinkExists                 = errors.New("link already tracked")
	ErrIncorrectRequestParameters = errors.New("incorrect request parameters")
	ErrChatAlreadyExists          = errors.New("chat already exists")
	ErrLinkNotExists              = errors.New("link not exists")
)

const requestTimeout = 5 * time.Second

type LinkRepository interface {
	AddLink(ctx context.Context, chatID int64, link pkg.LinkInfo) error
	DeleteLink(ctx context.Context, chatID int64, url string) (pkg.LinkInfo, error)
	GetUserLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error)
	GetAllLinks(ctx context.Context, host string, limit int, offset int) ([]pkg.LinkInfo, error)
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
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type LinksService struct {
	LinkRepo   LinkRepository
	ChatsRepo  ChatRepository
	Transactor Transactor
	BaseLogger *slog.Logger
}

func (h LinksService) AddChatID(ctx context.Context, chatID int64) error {
	return h.Transactor.WithTransaction(ctx, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
		if err != nil {
			return fmt.Errorf("error checking chat existence %w", err)
		}

		if exists {
			return ErrChatAlreadyExists
		}

		if err = h.ChatsRepo.AddChat(ctx, chatID); err != nil {
			return fmt.Errorf("error adding chat %w", err)
		}
		return nil
	})
}

func (h LinksService) DeleteChat(ctx context.Context, chatID int64) error {
	if err := h.ChatsRepo.DeleteChat(ctx, chatID); err != nil {
		return fmt.Errorf("error deleting chat %w", err)
	}
	return nil

}

func (h LinksService) GetLinks(ctx context.Context, chatID int64) ([]pkg.LinkInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("error checking chat existence %w", err)
	}
	if !exists {
		return nil, ErrChatNotFound
	}

	links, err := h.LinkRepo.GetUserLinks(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("error getting user links %w", err)
	}

	return links, nil
}

func (h LinksService) AddLink(ctx context.Context, chatID int64, linkRequest pkg.AddLinkRequest) error {
	return h.Transactor.WithTransaction(ctx, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		trackedLink := pkg.LinkInfo{Link: linkRequest.Link, Tags: linkRequest.Tags}
		if err := h.LinkRepo.AddLink(ctx, chatID, trackedLink); err != nil {
			return fmt.Errorf("error adding link %w", err)
		}
		return nil
	})

}

func (h LinksService) DeleteLink(ctx context.Context, chatID int64, link string) (pkg.LinkInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	exists, err := h.ChatsRepo.ChatExists(ctx, chatID)
	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("error checking chat existence %w", err)
	}
	if !exists {
		return pkg.LinkInfo{}, ErrChatNotFound
	}

	exists, err = h.LinkRepo.LinkExists(ctx, link)
	if err != nil {
		return pkg.LinkInfo{}, fmt.Errorf("error checking link %w", err)
	}
	if !exists {
		return pkg.LinkInfo{}, ErrLinkNotExists
	}

	linkInfo, err := h.LinkRepo.DeleteLink(ctx, chatID, link)
	if err != nil {
		slog.Error("error deleting link ", slog.String("error", err.Error()))
		return pkg.LinkInfo{}, err
	}
	return linkInfo, nil
}
