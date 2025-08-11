package web

type ListResponse struct {
	TotalResults int         `json:"total"`
	Page         int         `json:"page"`
	Limit        int         `json:"limit"`
	Results      interface{} `json:"results"`
}
