package bot

type LinkResponse struct {
	Filters []string `json:"filters,omitempty"`
	Id      int64    `json:"id,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Url     string   `json:"url,omitempty"`
}

type ListLinksResponse struct {
	Links []LinkResponse `json:"links,omitempty"`
	Size  int32          `json:"size,omitempty"`
}
