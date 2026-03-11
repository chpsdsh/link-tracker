package shared

import "time"

type LinkInfo struct {
	Link           string
	Tags           []string
	LastUpdateTime time.Time
}
