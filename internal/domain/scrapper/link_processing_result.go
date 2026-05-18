package scrapper

import (
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

type LinkProcessingResult struct {
	UpdateTime time.Time
	Events     []pkg.LinkUpdate
}
