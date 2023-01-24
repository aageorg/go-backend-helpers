package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type EndpointsAuthConfig interface {
	Schema(path string, accessType access_control.AccessType) (string, bool)
	DefaultSchema() string
}

type endpointSchema struct {
	ACCESS      access_control.AccessType
	HTTP_METHOD string
	SCHEMA      string
}

func (e *endpointSchema) Config() interface{} {
	return e
}

type EndpointsAuthConfigBaseConfig struct {
	DEFAULT_SCHEMA string `default:"jwt"`
}

type EndpointsAuthConfigBase struct {
	EndpointsAuthConfigBaseConfig
	endpoints map[string][]endpointSchema
}

func (e *EndpointsAuthConfigBase) Config() interface{} {
	return &e.EndpointsAuthConfigBaseConfig
}

func (e *EndpointsAuthConfigBase) Schema(path string, access access_control.AccessType) (string, bool) {

	ep, ok := e.endpoints[path]
	if !ok {
		return "", false
	}

	for _, epSchema := range ep {
		if access_control.Check(epSchema.ACCESS, access) {
			return epSchema.SCHEMA, true
		}
	}

	return "", false
}

func (e *EndpointsAuthConfigBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	path := utils.OptionalArg("endpoints_auth_config", configPath...)
	fields := logger.Fields{"where": "EndpointsAuthConfigBase.Init", "config_path": path}
	log.Info("Init configuration of endpoints authorization", fields)

	err := object_config.LoadLogValidate(cfg, log, vld, e, path)
	if err != nil {
		return log.Fatal("Failed to load configuration", err, fields)
	}

	e.endpoints = make(map[string][]endpointSchema)

	endpointsPath := object_config.Key(path, "endpoints")
	endpointsSection := cfg.Get(endpointsPath)
	endpoints := endpointsSection.(map[string]interface{})
	for endpoint := range endpoints {
		endpointPath := object_config.Key(path, endpoint)
		fields := utils.AppendMapNew(fields, logger.Fields{"endpoint": endpoint, "endpoint_path": endpointPath})
		endpointSchemas := make([]endpointSchema, 0)

		schemasSection := cfg.Get(endpointPath)
		schemas := schemasSection.([]interface{})
		for i := range schemas {
			schemaPath := object_config.KeyInt(endpointPath, i)
			fields := utils.AppendMapNew(fields, logger.Fields{"schema_path": schemaPath})
			epSchema := endpointSchema{}
			err := object_config.Load(cfg, schemaPath, &epSchema)
			if err != nil {
				return log.Fatal("Failed to load endpoint authorization schema", err, fields)
			}
			if epSchema.HTTP_METHOD != "" {
				epSchema.ACCESS = access_control.HttpMethod2Access(epSchema.HTTP_METHOD)
			}
			endpointSchemas = append(endpointSchemas, epSchema)
		}

		e.endpoints[endpoint] = endpointSchemas
	}

	// TODO log configuration

	return nil
}

func (e *EndpointsAuthConfigBase) DefaultSchema() string {
	return e.DEFAULT_SCHEMA
}
