// Code generated by go-swagger; DO NOT EDIT.

package schema

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// FailureResponse failure response
//
// swagger:model failureResponse
type FailureResponse struct {

	// data
	// Required: true
	Data interface{} `json:"data"`

	// status
	// Required: true
	// Enum: [fail]
	Status *string `json:"status"`
}

// Validate validates this failure response
func (m *FailureResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateData(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FailureResponse) validateData(formats strfmt.Registry) error {

	if m.Data == nil {
		return errors.Required("data", "body", nil)
	}

	return nil
}

var failureResponseTypeStatusPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["fail"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		failureResponseTypeStatusPropEnum = append(failureResponseTypeStatusPropEnum, v)
	}
}

const (

	// FailureResponseStatusFail captures enum value "fail"
	FailureResponseStatusFail string = "fail"
)

// prop value enum
func (m *FailureResponse) validateStatusEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, failureResponseTypeStatusPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *FailureResponse) validateStatus(formats strfmt.Registry) error {

	if err := validate.Required("status", "body", m.Status); err != nil {
		return err
	}

	// value enum
	if err := m.validateStatusEnum("status", "body", *m.Status); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this failure response based on context it is used
func (m *FailureResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *FailureResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FailureResponse) UnmarshalBinary(b []byte) error {
	var res FailureResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
