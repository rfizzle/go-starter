package web

type JSONStatus string

const (
	JSONStatusSuccess = JSONStatus("success")
	JSONStatusFail    = JSONStatus("fail")
	JSONStatusError   = JSONStatus("error")
)

type JSONResponse struct {
	Status  JSONStatus  `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code,omitempty"`
}
