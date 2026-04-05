package scrapper

import "time"

type GithubPullRequest struct {
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User struct {
		Login string `json:"login"`
	} `json:"user"`
}
