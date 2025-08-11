package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-faster/errors"

	"github.com/rfizzle/go-starter/internal/auth"
)

func UnauthenticatedRoute(_ auth.Authorization, contractFn Contract, endpointFn Endpoint) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defaultRouteLogic(ctx, contractFn, endpointFn, w, r)
	}
}

func AuthenticatedRoute(authFn auth.Authorization, contractFn Contract, endpointFn Endpoint) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Default context
		ctx := r.Context()

		// Handle nil authentication
		if authFn == nil {
			writeJSONError(marshalError(), w)
			return
		}

		// Authentication and authorization
		ctx, err := authFn(ctx, r)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "unauthenticated"):
				writeJSONError(&HTTPError{
					Code: http.StatusUnauthorized,
					Data: map[string]interface{}{"message": http.StatusText(http.StatusUnauthorized)},
				}, w)
			case strings.Contains(err.Error(), "unauthorized"):
				writeJSONError(&HTTPError{
					Code: http.StatusForbidden,
					Data: map[string]interface{}{"message": http.StatusText(http.StatusForbidden)},
				}, w)
			default:
				writeJSONError(&HTTPError{
					Code:    http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				}, w)
			}
			return
		}

		defaultRouteLogic(ctx, contractFn, endpointFn, w, r.WithContext(ctx))
	}
}

func defaultRouteLogic(ctx context.Context, contractFn Contract, endpointFn Endpoint, w http.ResponseWriter, r *http.Request) {
	// Parameter validation
	request, err := contractFn(r)
	if err != nil {
		writeJSONError(err, w)
		return
	}

	// Endpoint Logic
	response, err := endpointFn(ctx, request, w)
	if err != nil {
		writeJSONError(err, w)
		return
	}

	writeJSONResponse(response, w)
}

func writeJSONResponse(response *HTTPSuccess, w http.ResponseWriter) {
	// Default values
	responseCode := http.StatusOK
	var responseData interface{}
	responseHeaders := make(map[string]string)

	// Override default values if response is not nil
	if response != nil {
		responseCode = response.Code
		responseData = response.Data
		responseHeaders = response.Headers
	}

	// Write JSON response
	var jsonResponse JSONResponse
	jsonResponse.Status = JSONStatusSuccess
	jsonResponse.Code = responseCode
	jsonResponse.Data = responseData

	// Marshal JSON
	jsonBytes, err := json.Marshal(jsonResponse)
	if err != nil {
		writeJSONError(marshalError(), w)
		return
	}

	// Set default and additional headers
	setDefaultHeaders(w)
	setAdditionalHeaders(w, responseHeaders)

	// Write response
	w.WriteHeader(responseCode)
	_, _ = w.Write(jsonBytes)
}

func writeJSONError(jsonErr error, w http.ResponseWriter) {
	webError := mapError(jsonErr)
	var jsonResponse JSONResponse
	jsonResponse.Code = webError.Code
	if webError.Code > 299 && webError.Code < 500 {
		jsonResponse.Status = JSONStatusFail
		jsonResponse.Data = webError.Data
	} else {
		jsonResponse.Status = JSONStatusError
		jsonResponse.Message = webError.Message
	}

	jsonBody, err := json.Marshal(jsonResponse)

	if err != nil {
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(webError.Code)
		_, _ = w.Write(jsonBody)
	}
}

func marshalError() *HTTPError {
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Data:    nil,
	}
}

func setDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setAdditionalHeaders(w http.ResponseWriter, headers map[string]string) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
}

func mapError(err error) *HTTPError {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Data:    nil,
	}
}
