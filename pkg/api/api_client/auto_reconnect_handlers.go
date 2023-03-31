package api_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type AutoReconnectHandlers interface {
	GetRefreshToken() string
	SaveRefreshToken(ctx op_context.Context, token string)
	GetCredentials(ctx op_context.Context) (login string, password string, err error)
}
