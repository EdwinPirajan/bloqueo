package domain

type Config struct {
	Version string `json:"version"`
	Client  string `json:"client"`
}

type CheckResponse struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}
