package repository

import (
	"slices"
	"sync"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

type LinkRepository struct {
	linkStorage map[int64][]shared.LinkInfo
	mu          sync.RWMutex
}

func NewLinkRepository() *LinkRepository {
	return &LinkRepository{linkStorage: make(map[int64][]shared.LinkInfo), mu: sync.RWMutex{}}
}

func (r *LinkRepository) ChatExists(chatID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.linkStorage[chatID]
	return ok
}

func (r *LinkRepository) AddChat(chatID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatID] = []shared.LinkInfo{}
}

func (r *LinkRepository) GetLinks(chatID int64) []shared.LinkInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.linkStorage[chatID]
}

func (r *LinkRepository) AddLink(chatID int64, link shared.LinkInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatID] = append(r.linkStorage[chatID], link)
}

func (r *LinkRepository) DeleteLink(chatID int64, link string) (shared.LinkInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, l := range r.linkStorage[chatID] {
		if l.Link == link {
			r.linkStorage[chatID] = slices.Delete(r.linkStorage[chatID], i, i+1)
			return l, true
		}
	}
	return shared.LinkInfo{}, false
}

func (r *LinkRepository) DeleteChat(chatID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.linkStorage, chatID)
}

func (r *LinkRepository) GetAllLinks() []shared.LinkInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	links := make([]shared.LinkInfo, 0)
	for _, l := range r.linkStorage {
		links = append(links, l...)
	}
	return links
}

func (r *LinkRepository) GetChatIDsByLink(link string) []int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]int64, 0)

	for id, chatLinks := range r.linkStorage {
		for _, l := range chatLinks {
			if l.Link == link {
				ids = append(ids, id)
				break
			}
		}
	}

	return ids
}

func (r *LinkRepository) UpdateLinksTime(newTime time.Time, linkToUpdate shared.LinkInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, chatLinks := range r.linkStorage {
		for i, link := range chatLinks {
			if link.Link == linkToUpdate.Link {
				chatLinks[i].LastUpdateTime = newTime
				break
			}
		}
	}
}
