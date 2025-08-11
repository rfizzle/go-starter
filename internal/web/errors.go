package web

import "net/http"

func invalidRequestWithMessage(message string) *HTTPError {
	return &HTTPError{
		Code: http.StatusBadRequest,
		Data: map[string]interface{}{"message": message},
	}
}

func invalidRequestWithMessageAndErrors(message string, errors []string) *HTTPError {
	return &HTTPError{
		Code: http.StatusBadRequest,
		Data: map[string]interface{}{
			"message": message,
			"errors":  errors,
		},
	}
}
