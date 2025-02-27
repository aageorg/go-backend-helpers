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
}

type endpointSchema struct {
	ACCESS      access_control.AccessType
	HTTP_METHOD string
	SCHEMA      string
}

func (e *endpointSchema) Config() interface{} {
	return e
}

type EndpointsAuthConfigBase struct {
	endpoints map[string][]endpointSchema
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
	fields := logger.Fields{"config_path": path}
	log.Debug("Init configuration of endpoints authorization", fields)

	e.endpoints = make(map[string][]endpointSchema)

	endpointsSection := cfg.Get(path)
	endpoints := endpointsSection.(map[string]interface{})
	for endpoint := range endpoints {
		endpointPath := object_config.Key(path, endpoint)
		fields := utils.AppendMapNew(fields, logger.Fields{"endpoint": endpoint, "endpoint_path": endpointPath})
		endpointSchemas := make([]endpointSchema, 0)

		log.Debug("Add auth schemas for endpoint", fields)

		schemasSection := cfg.Get(endpointPath)
		schemas := schemasSection.([]interface{})
		for i := range schemas {
			schemaPath := object_config.KeyInt(endpointPath, i)
			fields := utils.AppendMapNew(fields, logger.Fields{"schema_path": schemaPath})
			epSchema := endpointSchema{}
			err := object_config.Load(cfg, schemaPath, &epSchema)
			if err != nil {
				return log.PushFatalStack("failed to load endpoint authorization schema", err, fields)
			}
			fields["access"] = epSchema.ACCESS
			fields["http_method"] = epSchema.HTTP_METHOD
			fields["schema"] = epSchema.SCHEMA
			if epSchema.HTTP_METHOD != "" {
				epSchema.ACCESS = access_control.HttpMethod2Access(epSchema.HTTP_METHOD)
			}
			endpointSchemas = append(endpointSchemas, epSchema)

			log.Info("Add auth schema", fields)
		}

		e.endpoints[endpoint] = endpointSchemas
	}

	return nil
}
