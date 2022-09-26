package requests

type Resize struct {
	URLs   []string `json:"urls"`
	Width  uint     `json:"width"`
	Height uint     `json:"height"`
}
