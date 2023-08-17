// Code generated by go-swagger; DO NOT EDIT.

package health

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// HealthLivenessReader is a Reader for the HealthLiveness structure.
type HealthLivenessReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *HealthLivenessReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewHealthLivenessOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, runtime.NewAPIError("[GET /healthz/liveness] HealthLiveness", response, response.Code())
	}
}

// NewHealthLivenessOK creates a HealthLivenessOK with default headers values
func NewHealthLivenessOK() *HealthLivenessOK {
	return &HealthLivenessOK{}
}

/*
HealthLivenessOK describes a response with status code 200, with default header values.

success
*/
type HealthLivenessOK struct {
}

// IsSuccess returns true when this health liveness o k response has a 2xx status code
func (o *HealthLivenessOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this health liveness o k response has a 3xx status code
func (o *HealthLivenessOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this health liveness o k response has a 4xx status code
func (o *HealthLivenessOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this health liveness o k response has a 5xx status code
func (o *HealthLivenessOK) IsServerError() bool {
	return false
}

// IsCode returns true when this health liveness o k response a status code equal to that given
func (o *HealthLivenessOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the health liveness o k response
func (o *HealthLivenessOK) Code() int {
	return 200
}

func (o *HealthLivenessOK) Error() string {
	return fmt.Sprintf("[GET /healthz/liveness][%d] healthLivenessOK ", 200)
}

func (o *HealthLivenessOK) String() string {
	return fmt.Sprintf("[GET /healthz/liveness][%d] healthLivenessOK ", 200)
}

func (o *HealthLivenessOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
