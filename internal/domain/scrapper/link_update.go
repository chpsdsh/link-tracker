package scrapper

import "time"

type linkUpdate struct {
	UpdatedAt time.Time `json:"updated_at"`
}
