package web

import (
	"fmt"
	"strings"
)

type HTTPSuccess struct {
	Code    int               `json:"-"`
	Data    interface{}       `json:"-"`
	Headers map[string]string `json:"-"`
}

type HTTPError struct {
	Code    int                    `json:"-"`
	Message string                 `json:"-"`
	Data    map[string]interface{} `json:"-"`
	Err     error                  `json:"err"`
}

func (err *HTTPError) Error() string {
	if err.Code < 500 {
		errString := strings.Builder{}
		for k, v := range err.Data {
			errString.WriteString(fmt.Sprintf("%s: %v;", k, v))
		}
		return errString.String()
	}
	return err.Message
}
