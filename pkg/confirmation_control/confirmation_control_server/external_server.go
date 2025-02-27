package confirmation_control_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/api/pool_microservice/pool_misrocervice_client"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api/confirmation_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api/confirmation_api_service"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ExternalServerConfig struct {
	EXPLICIT_CODE_CHECK bool
}

type ExternalServer struct {
	ExternalServerConfig

	auth         auth.Auth
	server       api_server.Server
	smsManager   sms.SmsManager
	smsProviders sms.ProviderFactory
}

type ExternalServerCfg struct {
	SmsManager   sms.SmsManager
	SmsProviders sms.ProviderFactory
}

func (s *ExternalServer) Construct(config ...ExternalServerCfg) {
	if len(config) != 0 {
		cfg := config[0]
		s.smsManager = cfg.SmsManager
		s.smsProviders = cfg.SmsProviders
	}
}

func NewExternalServer(config ...ExternalServerCfg) *ExternalServer {
	s := &ExternalServer{}
	s.Construct(config...)
	return s
}

func (s *ExternalServer) Config() interface{} {
	return &s.ExternalServerConfig
}

func (s *ExternalServer) Init(app app_with_multitenancy.AppWithMultitenancy, configPath ...string) error {

	path := utils.OptionalArg("external_server", configPath...)

	err := object_config.LoadLogValidate(app.Cfg(), app.Logger(), app.Validator(), s, path)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init external server of confirmation control server", err)
	}

	// init SMS manager
	if s.smsManager == nil && s.smsProviders != nil {
		smsManager := sms.NewSmsManager()
		err := smsManager.Init(app.Cfg(), app.Logger(), app.Validator(), s.smsProviders, "sms")
		if err != nil {
			return app.Logger().PushFatalStack("failed to init SMS manager", err)
		}
		s.smsManager = smsManager
	}

	// init auth controller
	if s.auth == nil {
		auth := auth.NewAuth()
		authPath := object_config.Key(path, "auth")
		err := auth.Init(app.Cfg(), app.Logger(), app.Validator(), &AuthFactory{SmsManager: s.smsManager}, authPath)
		if err != nil {
			return app.Logger().PushFatalStack("failed to init auth manager", err)
		}
		s.auth = auth
	}

	// init REST API server
	server := rest_api_gin_server.NewServer()
	serverPath := object_config.Key(path, "rest_api_server")
	err = server.Init(app, s.auth, app.Multitenancy(), serverPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init REST API server", err)
	}
	s.server = server

	// create callback client
	callbackTransportPath := object_config.Key(path, "callback_client")
	callbackTransport := pool_misrocervice_client.NewPoolMicroserviceClient("confirmation_callback")
	err = callbackTransport.Init(app, callbackTransportPath)
	if err != nil {
		return app.Logger().PushFatalStack("failed to init callback client", err)
	}
	callbackClient := confirmation_api_client.NewConfirmationCallbackClient(callbackTransport)

	// add services
	externalService := confirmation_api_service.NewConfirmationExternalService(callbackClient, s.EXPLICIT_CODE_CHECK)
	api_server.AddServiceToServer(s.server, externalService)

	// done
	return nil
}

func (s *ExternalServer) ApiServer() api_server.Server {
	return s.server
}
