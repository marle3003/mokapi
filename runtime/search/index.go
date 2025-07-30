package search

type Result struct {
	Results []ResultItem `json:"results"`
	Total   uint64       `json:"total"`
}

type ResultItem struct {
	Type      string            `json:"type"`
	Domain    string            `json:"domain,omitempty"`
	Title     string            `json:"title"`
	Fragments []string          `json:"fragments,omitempty"`
	Params    map[string]string `json:"params"`
	Time      string            `json:"time,omitempty"`
}

type ErrNotEnabled struct{}

func (e *ErrNotEnabled) Error() string {
	return "search is not enabled"
}

type Request struct {
	QueryText string
	Index     int
	Limit     int
}
