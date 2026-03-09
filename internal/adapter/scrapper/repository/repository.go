package repository

import (
	"slices"
	"sync"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

type LinkRepository struct {
	linkStorage map[int64][]shared.LinkInfo
	mu          sync.RWMutex
}

func NewLinkRepository() *LinkRepository {
	return &LinkRepository{linkStorage: make(map[int64][]shared.LinkInfo), mu: sync.RWMutex{}}
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
	r.linkStorage[chatId] = []shared.LinkInfo{}
}

func (r *LinkRepository) GetLinks(chatId int64) []shared.LinkInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.linkStorage[chatId]
}

func (r *LinkRepository) AddLink(chatId int64, link shared.LinkInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatId] = append(r.linkStorage[chatId], link)
}

func (r *LinkRepository) DeleteLink(chatId int64, link string) (shared.LinkInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, l := range r.linkStorage[chatId] {
		if l.Link == link {
			r.linkStorage[chatId] = slices.Delete(r.linkStorage[chatId], i, i+1)
			return l, true
		}
	}
	return shared.LinkInfo{}, false
}

func (r *LinkRepository) DeleteChat(chatId int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.linkStorage, chatId)
}
