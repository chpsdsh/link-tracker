package service

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

type LinksPool struct {
	LinksChan  chan []pkg.LinkInfo
	NumWorkers int
}

func NewLinksPool(numWorkers int) LinksPool {
	linksChan := make(chan []pkg.LinkInfo)
	return LinksPool{
		LinksChan:  linksChan,
		NumWorkers: numWorkers,
	}
}
