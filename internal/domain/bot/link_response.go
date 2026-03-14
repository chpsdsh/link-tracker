package bot

type LinkResponse struct {
	Filters []string `json:"filters,omitempty"`
	ID      int64    `json:"id,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	URL     string   `json:"url,omitempty"`
}

type ListLinksResponse struct {
	Links []LinkResponse `json:"links,omitempty"`
	Size  int32          `json:"size,omitempty"`
}
