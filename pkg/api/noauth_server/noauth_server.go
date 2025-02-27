package noauth_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Server interface {
	ApiServer() api_server.Server
	Auth() auth.Auth
}

type NoAuthServer struct {
	auth   auth.Auth
	server api_server.Server
}

type Config struct {
	Server api_server.Server
}

func New(config ...Config) *NoAuthServer {
	s := &NoAuthServer{}
	s.Construct(config...)
	return s
}

func (s *NoAuthServer) Construct(config ...Config) {
	if len(config) != 0 {
		cfg := config[0]
		s.server = cfg.Server
	}
}

func (s *NoAuthServer) Init(app app_with_multitenancy.AppWithMultitenancy, configPath ...string) error {

	path := utils.OptionalArg("server", configPath...)

	// noauth
	s.auth = auth.NewNoAuth()

	// init REST API server
	if s.server == nil {
		server := rest_api_gin_server.NewServer()
		serverPath := object_config.Key(path, "rest_api_server")
		err := server.Init(app, s.auth, app.Multitenancy(), serverPath)
		if err != nil {
			return app.Logger().PushFatalStack("failed to init REST API server", err)
		}
		s.server = server
	}

	// done
	return nil
}

func (s *NoAuthServer) Auth() auth.Auth {
	return s.auth
}

func (s *NoAuthServer) ApiServer() api_server.Server {
	return s.server
}
