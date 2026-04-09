package scrapper

import "time"

type GitHubRepositoryResponse struct {
	UpdatedAt time.Time `json:"updated_at"`
}
