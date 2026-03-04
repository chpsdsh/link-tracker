package shared

import "net/url"

type TrackedURL struct {
	Url  url.URL
	Tags []string
}
