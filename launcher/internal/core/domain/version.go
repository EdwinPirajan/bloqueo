package domain

type Config struct {
	Version string `json:"version"`
}

type CheckResponse struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	URL     string `json:"url"`
}
