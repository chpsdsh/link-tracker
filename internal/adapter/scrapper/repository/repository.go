package repository

import (
	"slices"
	"sync"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

type LinkRepository struct {
	linkStorage map[int64][]shared.TrackedLink
	mu          sync.RWMutex
}

func NewLinkRepository() *LinkRepository {
	return &LinkRepository{linkStorage: make(map[int64][]shared.TrackedLink), mu: sync.RWMutex{}}
}

func (r *LinkRepository) ChatExists(chatId int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.linkStorage[chatId]
	return ok
}

func (r *LinkRepository) AddChat(chatId int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatId] = []shared.TrackedLink{}
}

func (r *LinkRepository) GetLinks(chatId int64) []shared.TrackedLink {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.linkStorage[chatId]
}

func (r *LinkRepository) AddLink(chatId int64, link shared.TrackedLink) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatId] = append(r.linkStorage[chatId], link)
}

func (r *LinkRepository) DeleteLink(chatId int64, link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.linkStorage[chatId] {
		if u.Link == link {
			r.linkStorage[chatId] = slices.Delete(r.linkStorage[chatId], i, i+1)
			break
		}
	}
}
