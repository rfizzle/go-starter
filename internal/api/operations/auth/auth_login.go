// Code generated by go-swagger; DO NOT EDIT.

package auth

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/rfizzle/go-starter/internal/entity"
)

// AuthLoginHandlerFunc turns a function with the right signature into a auth login handler
type AuthLoginHandlerFunc func(AuthLoginParams, entity.Entity) middleware.Responder

// Handle executing the request and returning a response
func (fn AuthLoginHandlerFunc) Handle(params AuthLoginParams, principal entity.Entity) middleware.Responder {
	return fn(params, principal)
}

// AuthLoginHandler interface for that can handle valid auth login params
type AuthLoginHandler interface {
	Handle(AuthLoginParams, entity.Entity) middleware.Responder
}

// NewAuthLogin creates a new http.Handler for the auth login operation
func NewAuthLogin(ctx *middleware.Context, handler AuthLoginHandler) *AuthLogin {
	return &AuthLogin{Context: ctx, Handler: handler}
}

/*
	AuthLogin swagger:route POST /api/v1/auth/login auth authLogin

# Login a user

Authenticates a user from a username and password and returns a JWT in the response and inside a
signed cookie.
*/
type AuthLogin struct {
	Context *middleware.Context
	Handler AuthLoginHandler
}

func (o *AuthLogin) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewAuthLoginParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal entity.Entity
	if uprinc != nil {
		principal = uprinc.(entity.Entity) // this is really a entity.Entity, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
