// Code generated by go-swagger; DO NOT EDIT.

package health

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// HealthReadinessHandlerFunc turns a function with the right signature into a health readiness handler
type HealthReadinessHandlerFunc func(HealthReadinessParams) middleware.Responder

// Handle executing the request and returning a response
func (fn HealthReadinessHandlerFunc) Handle(params HealthReadinessParams) middleware.Responder {
	return fn(params)
}

// HealthReadinessHandler interface for that can handle valid health readiness params
type HealthReadinessHandler interface {
	Handle(HealthReadinessParams) middleware.Responder
}

// NewHealthReadiness creates a new http.Handler for the health readiness operation
func NewHealthReadiness(ctx *middleware.Context, handler HealthReadinessHandler) *HealthReadiness {
	return &HealthReadiness{Context: ctx, Handler: handler}
}

/*
	HealthReadiness swagger:route GET /healthz/readiness health healthReadiness

# Readiness probe

Readiness probe for kubernetes health check. Returns 200 if the service is ready to serve requests.
Returns 503 if the service is not ready to serve requests (starting up or shutting down).
*/
type HealthReadiness struct {
	Context *middleware.Context
	Handler HealthReadinessHandler
}

func (o *HealthReadiness) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewHealthReadinessParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
