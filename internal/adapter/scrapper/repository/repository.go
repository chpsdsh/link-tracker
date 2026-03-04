package repository

import (
	"net/url"
	"slices"
	"sync"
)

type LinkRepository struct {
	linkStorage map[int64][]url.URL
	mu          sync.RWMutex
}

func (r *LinkRepository) GetLinks(chatId int64) []url.URL {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.linkStorage[chatId]
}

func (r *LinkRepository) AddLink(chatId int64, link url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkStorage[chatId] = append(r.linkStorage[chatId], link)
}

func (r *LinkRepository) DeleteLink(chatId int64, link url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, u := range r.linkStorage[chatId] {
		if u.String() == link.String() {
			r.linkStorage[chatId] = slices.Delete(r.linkStorage[chatId], i, i+1)
			break
		}
	}
}
