package server

import (
	"fmt"
	"net/http"

	"github.com/rfizzle/go-starter/internal/entity"

	"github.com/rfizzle/go-starter/internal/api"
	"github.com/rfizzle/go-starter/internal/controller"
	"go.uber.org/zap"
)

type Server struct {
	//httpServer *http.Server
	logger  *zap.Logger
	handler http.Handler
}

func New(logger *zap.Logger) (*Server, error) {
	apiHandler, err := setupApi(logger, nil)
	if err != nil {
		return nil, err
	}

	return &Server{
		handler: apiHandler,
		logger:  logger,
	}, nil
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Stop() error {
	return nil
}

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
