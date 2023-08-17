package server

import (
	"fmt"
	"net/http"

	"github.com/rfizzle/go-starter/internal/api"
	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/entity"
	"go.uber.org/zap"
)

func setupApi(logger *zap.Logger, controller *controller.Controller) (http.Handler, error) {
	httpHandler, err := api.Handler(api.Config{
		HealthAPI: controller.Health,
		Logger:    logger.Named("api").Sugar().Infof,
		//InnerMiddleware:     nil,
		//Authorizer:          nil,
		AuthHasPermission: func(token string, scopes []string) (entity.Entity, error) {
			// TODO: Take token and lookup entity in database
			// Populate it with the scopes/permissions and compare to the supplied scopes
			// If the entity has the required scopes, return it
			// If the entity does not have the required scopes, return nil and an error
			if token == "1234567890" {
				return entity.NewAuthEntity("1234567890"), nil
			}
			return nil, fmt.Errorf("unauthorized")
		},
		//APIKeyAuthenticator: nil,
		//BasicAuthenticator:  nil,
		//BearerAuthenticator: nil,
	})
	if err != nil {
		return nil, err
	}

	return httpHandler, nil
}
