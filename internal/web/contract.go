package web

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
	"github.com/gookit/validate"
	"github.com/pasztorpisti/qs"
)

func init() {
	// Setup validators
	validate.AddGlobalMessages(map[string]string{
		"required": "{field} is required",
		"eqField":  "{field} must be equal to %s",
	})
}

type Contract func(r *http.Request) (interface{}, error)

func EmptyContract() Contract {
	return func(r *http.Request) (interface{}, error) {
		return nil, nil
	}
}

func _contractRequestBodyUnmarshal(r *http.Request, requestObject interface{}) *HTTPError {
	var body []byte

	if r.Body != nil {
		// Get request body
		tmpBody, err := io.ReadAll(r.Body)
		if err != nil {
			return invalidRequestWithMessage("issue reading json body")
		}
		body = tmpBody
	}

	// If the body is not null, try to marshal
	if len(body) > 0 {
		if err := json.Unmarshal(body, requestObject); err != nil {
			return invalidRequestWithMessage("issue decoding json body")
		}
	}

	// Get query params
	requestValues := r.URL.Query()

	// Get path params from chi
	routeContext := chi.RouteContext(r.Context())
	if routeContext != nil {
		urlParamKeys := routeContext.URLParams.Keys
		for _, key := range urlParamKeys {
			requestValues.Set(key, routeContext.URLParam(key))
		}
	}

	// Marshal url param values
	if err := qs.UnmarshalValues(requestObject, requestValues); err != nil {
		return invalidRequestWithMessage("issue decoding url params")
	}

	return nil
}

func _contractLogic(requestObject interface{}) *HTTPError {
	if reflect.ValueOf(requestObject).Kind() != reflect.Ptr {
		return invalidRequestWithMessage("request object must be a pointer")
	}

	// Validate structure
	v := validate.Struct(requestObject)
	if v.Validate() {
		err := v.BindSafeData(requestObject)
		if err != nil {
			return invalidRequestWithMessage("invalid structure")
		}
		return nil
	}
	if !v.Errors.Empty() {
		mappedErrors := mapContractErrorsToArray(v.Errors.All())
		return invalidRequestWithMessageAndErrors("parameter validation failed", mappedErrors)
	}
	return invalidRequestWithMessage("contract error")
}

func RequestContractLogic(r *http.Request, requestObject interface{}) *HTTPError {
	if reflect.ValueOf(requestObject).Kind() != reflect.Ptr {
		return invalidRequestWithMessage("request object must be a pointer")
	}
	err := _contractRequestBodyUnmarshal(r, requestObject)
	if err != nil {
		return err
	}
	return _contractLogic(requestObject)
}

func mapContractErrorsToArray(err map[string]map[string]string) []string {
	arr := make([]string, 0)
	for _, v := range err {
		for _, v2 := range v {
			arr = append(arr, v2)
		}
	}
	return arr
}
