package responses

type Resize struct {
	Result string `json:"result"`
	URL    string `json:"url,omitempty"`
	Cached bool   `json:"cached"`
}
