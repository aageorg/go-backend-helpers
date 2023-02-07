package rest_api_gin_server

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_csrf"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gin-gonic/gin"

	finish "github.com/evgeniums/go-finish-service"
)

const TenancyParameter string = "tenancy"

type ServerConfig struct {
	api_server.ServerBaseConfig

	HOST                    string `validate:"ip" default:"127.0.0.1"`
	PORT                    uint16 `validate:"required"`
	PATH_PREFIX             string `default:"/api"`
	TRUSTED_PROXIES         []string
	VERBOSE                 bool
	VERBOSE_BODY_MAX_LENGTH int `default:"2048"`
}

type AuthParameterGetter = func(r *Request, key string) string
type AuthParameterSetter = func(r *Request, key string, value string)

type Server struct {
	ServerConfig
	multitenancy.MultitenancyBase
	app_context.WithAppBase
	generic_error.ErrorManagerBaseHttp
	auth.WithAuthBase

	ginEngine     *gin.Engine
	notFoundError *api.ResponseError
	hostname      string

	authParamsGetters map[string]AuthParameterGetter
	authParamsSetters map[string]AuthParameterSetter

	csrf            *auth_csrf.AuthCsrf
	tenancyResource api.Resource
}

func getHttpHeader(g *gin.Context, name string) string {
	return g.GetHeader(name)
}

func NewServer() *Server {
	s := &Server{}

	csrfKey := func(key string) string {
		return utils.ConcatStrings("x-", key)
	}

	s.authParamsSetters = make(map[string]AuthParameterSetter, 0)
	s.authParamsSetters[auth_csrf.AntiCsrfProtocol] = func(r *Request, key string, value string) {
		r.ginCtx.Header(csrfKey(key), value)
	}

	s.authParamsGetters = make(map[string]AuthParameterGetter, 0)
	s.authParamsGetters[auth_csrf.AntiCsrfProtocol] = func(r *Request, key string) string {
		name := csrfKey(key)
		return getHttpHeader(r.ginCtx, name)
	}

	return s
}

func (s *Server) Config() interface{} {
	return &s.ServerConfig
}

func (s *Server) Testing() bool {
	return s.App().Testing()
}

func (s *Server) address() string {
	a := fmt.Sprintf("%s:%d", s.HOST, s.PORT)
	return a
}

func (s *Server) logGinRequest(log logger.Logger, path string, start time.Time, ginCtx *gin.Context, extraFields ...logger.Fields) {

	stop := time.Since(start)
	latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
	statusCode := ginCtx.Writer.Status()
	clientIP := ginCtx.ClientIP()
	clientUserAgent := ginCtx.Request.UserAgent()
	referer := ginCtx.Request.Referer()
	dataLength := ginCtx.Writer.Size()
	if dataLength < 0 {
		dataLength = 0
	}

	msg := "HTTP request"
	fields := logger.Fields{
		"hostname":    s.hostname,
		"http_code":   statusCode,
		"latency":     latency, // time to process
		"client_ip":   clientIP,
		"method":      ginCtx.Request.Method,
		"path":        path,
		"referer":     referer,
		"data_length": dataLength,
		"user_agent":  clientUserAgent,
	}
	logger.AppendFields(fields, extraFields...)

	if len(ginCtx.Errors) > 0 {
		log.Error(msg, errors.New(ginCtx.Errors.ByType(gin.ErrorTypePrivate).String()), fields)
	} else {
		if statusCode >= http.StatusInternalServerError {
			log.Error(msg, errors.New("internal server error"), fields)
		} else if statusCode >= http.StatusBadRequest {
			log.Warn(msg, fields)
		} else {
			log.Info(msg, fields)
		}
	}

	ginCtx.Set("logged", true)
}

func (s *Server) ginDefaultLogger() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {

		path := ginCtx.Request.URL.Path
		start := time.Now()

		ginCtx.Next()

		// skip if request was already logged
		_, logged := ginCtx.Get("logged")
		if logged {
			return
		}

		s.logGinRequest(s.App().Logger(), path, start, ginCtx, logger.Fields{"status": s.notFoundError.Code})
	}
}

func (s *Server) NoRoute() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ginCtx.JSON(http.StatusNotFound, s.notFoundError)
	}
}

func (s *Server) Init(ctx app_context.Context, auth auth.Auth, configPath ...string) error {

	var err error
	s.hostname, err = os.Hostname()
	if err != nil {
		s.hostname = "unknow"
	}
	ctx.Logger().Info("Init REST API gin server", logger.Fields{"hostname": s.hostname})

	s.WithAppBase.Init(ctx)
	s.ErrorManagerBaseHttp.Init()
	s.WithAuthBase.Init(auth)
	auth.AttachToErrorManager(s)

	if s.IsMultiTenancy() {
		s.tenancyResource = api.NewResource(TenancyParameter, api.ResourceConfig{HasId: true})
	}

	defaultPath := "rest_api_server"

	// load configuration
	err = object_config.LoadLogValidate(ctx.Cfg(), ctx.Logger(), ctx.Validator(), s, defaultPath, configPath...)
	if err != nil {
		return ctx.Logger().PushFatalStack("failed to load server configuration", err, logger.Fields{"name": s.Name()})
	}

	// load CSRF configuration
	csrfKey := object_config.Key(utils.OptionalArg(defaultPath, configPath...), "csrf")
	if ctx.Cfg().IsSet(csrfKey) {
		s.csrf = auth_csrf.New()
		err = s.csrf.Init(ctx.Cfg(), ctx.Logger(), ctx.Validator(), csrfKey)
		if err != nil {
			return ctx.Logger().PushFatalStack("failed to load anti-CSRF configuration", err)
		}

		s.AddErrorDescriptions(s.csrf.ErrorDescriptions())
		s.AddErrorProtocolCodes(s.csrf.ErrorProtocolCodes())
	}

	// init gin router
	s.ginEngine = gin.New()
	// trusted proxies are needed for correct logging of client IP address
	s.ginEngine.SetTrustedProxies(s.TRUSTED_PROXIES)
	// use default logger for unhandled paths, use recovery middleware to catch panic failures
	s.ginEngine.Use(s.ginDefaultLogger(), gin.Recovery())

	// set noroute
	s.notFoundError = &api.ResponseError{Code: "not_found", Message: "Requested resource was not found"}
	s.ginEngine.NoRoute(s.NoRoute())

	// done
	return nil
}

func (s *Server) Run(fin *finish.Finisher) {

	srv := &http.Server{Addr: s.address(), Handler: s.ginEngine}
	fin.Add(srv)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			msg := "failed to start HTTP server"
			fmt.Printf("%s %s: %s\n", msg, s.Name(), err)
			s.App().Logger().Fatal(msg, err, logger.Fields{"name": s.Name()})
		}
	}()
}

func requestHandler(s *Server, ep api_server.Endpoint) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {

		var err error

		// create and init request
		request := &Request{}
		request.Init(s, ginCtx, ep)
		request.SetName(ep.Name())

		if s.VERBOSE {
			dumpBody := ginCtx.Request.ContentLength > 0 && int(ginCtx.Request.ContentLength) <= s.VERBOSE_BODY_MAX_LENGTH
			b, _ := httputil.DumpRequest(ginCtx.Request, dumpBody)
			request.Logger().Debug("Dump HTTP request", logger.Fields{"request": string(b)})
		}

		// extract tenancy if applicable
		if s.IsMultiTenancy() {
			tenancyInPath := request.GetResourceId(TenancyParameter)
			tenancy, err := s.TenancyByPath(tenancyInPath)
			if err != nil {
				// report that tenancy was not found
				request.SetGenericError(s.MakeGenericError(generic_error.ErrorCodeNotFound, request.Tr))
				request.Logger().ErrorNative(err, logger.Fields{"tenancy": tenancyInPath})
			} else {
				request.SetTenancy(tenancy)
			}
		}

		// process CSRF
		if err == nil {
			if s.csrf != nil {
				_, err = s.csrf.Handle(request)
			}
		}

		// process auth
		if err == nil {
			err = s.Auth().HandleRequest(request, ep.Resource().ServicePathPrototype(), ep.AccessType())
			if err != nil {
				request.SetGenericError(s.MakeGenericError(auth.ErrorCodeUnauthorized, request.Tr))
				// errors must be processed in handler
			}
		}

		// TODO process access control
		if err == nil {
			// errors must be processed in handler
		}

		// call endpoint's request handler
		if err == nil {
			ep.HandleRequest(request) // do we need to handle error return here?
		}

		// close context with sending response to client
		request.Close()
	}
}

func (s *Server) AddEndpoint(ep api_server.Endpoint) {

	if ep.TestOnly() && !s.Testing() {
		return
	}

	ep.AttachToErrorManager(s)

	method := access_control.Access2HttpMethod(ep.AccessType())
	if method == "" {
		panic(fmt.Sprintf("Invalid HTTP method in endpoint %s for access %d", ep.Name(), ep.AccessType()))
	}

	if s.IsMultiTenancy() {
		s.tenancyResource.AddChild(ep.Resource())
	}

	path := fmt.Sprintf("%s/%s%s", s.PATH_PREFIX, s.ApiVersion(), ep.Resource().FullPathPrototype())
	s.ginEngine.Handle(method, path, requestHandler(s, ep))
}

func (s *Server) MakeResponseError(gerr generic_error.Error) (int, *api.ResponseError) {
	err := &api.ResponseError{Code: gerr.Code(), Message: gerr.Message(), Details: gerr.Details()}
	code := s.ErrorProtocolCode(gerr.Code())
	return code, err
}

func (s *Server) AuthParameterGetter(authMethodProtocol string) AuthParameterGetter {
	handler, ok := s.authParamsGetters[authMethodProtocol]
	if !ok {
		return nil
	}
	return handler
}

func (s *Server) AuthParameterSetter(authMethodProtocol string) AuthParameterSetter {
	handler, ok := s.authParamsSetters[authMethodProtocol]
	if !ok {
		return nil
	}
	return handler
}

func (s *Server) GinEngine() *gin.Engine {
	return s.ginEngine
}
