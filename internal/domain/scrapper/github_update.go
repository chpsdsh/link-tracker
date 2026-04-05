package scrapper

import "time"

type GitHubUpdate struct {
	UpdatedAt time.Time `json:"updated_at"`
}
