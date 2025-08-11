package driver

import (
	"go.uber.org/zap"
)

type Provider interface {
	Name() string
	Ping() error
	Disconnect() error
	Logger() *zap.Logger

	WithTx(fn func(tx Tx) error) error
	Client
}
