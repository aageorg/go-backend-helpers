package object_config

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type WithInit interface {
	Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error
}

type WithProtocol interface {
	Protocol() string
}

type WithProtocolBase struct {
	PROTOCOL string `gorm:"index" json:"protocol" validate:"required" long:"protocol" description:"Protocol"`
}

func (p *WithProtocolBase) Protocol() string {
	return p.PROTOCOL
}

type Subobject interface {
	WithProtocol
	common.WithName
	WithInit
}

type SubobjectFactory[T Subobject] func(protocol string) (T, error)

func LoadLogValidateSubobjectsMap[T Subobject](cfg config.Config, log logger.Logger, vld validator.Validator, configPath string, createSubobjectFnc SubobjectFactory[T], loggerFields ...logger.Fields) (map[string]T, error) {

	if !cfg.IsSet(configPath) {
		return nil, nil
	}

	fields := logger.AppendFieldsNew(logger.Fields{"config_path": configPath}, loggerFields...)
	subobjects := make(map[string]T)

	subobjectsSection := cfg.Get(configPath)
	subobjectsConfig, ok := subobjectsSection.(map[string]interface{})
	if !ok {
		return nil, log.PushFatalStack("invalid subobjects section in configuration", nil, fields)
	}
	for subobjectName := range subobjectsConfig {
		path := Key(configPath, subobjectName)
		protocolPath := Key(path, "protocol")
		protocol := cfg.GetString(protocolPath)
		fields := utils.AppendMapNew(fields, logger.Fields{"name": subobjectName, "config_path": path, "protocol": protocol})
		subobject, err := createSubobjectFnc(protocol)
		if err != nil {
			return nil, log.PushFatalStack("failed to create subobject", err, fields)
		}
		err = subobject.Init(cfg, log, vld, path)
		if err != nil {
			return nil, log.PushFatalStack("failed to initialize subobject", err, fields)
		}
		subobject.SetName(subobjectName)
		subobjects[subobjectName] = subobject
	}

	return subobjects, nil
}

type SubobjectBuilder[T WithInit] func() T

func LoadLogValidateSubobjectsList[T WithInit](cfg config.Config, log logger.Logger, vld validator.Validator, configPath string, createSubobjectFnc SubobjectBuilder[T], loggerFields ...logger.Fields) ([]T, error) {

	if !cfg.IsSet(configPath) {
		return nil, nil
	}

	fields := logger.AppendFieldsNew(logger.Fields{"config_path": configPath}, loggerFields...)
	subobjects := make([]T, 0)

	subobjectsSection := cfg.Get(configPath)
	subobjectsConfig, ok := subobjectsSection.([]interface{})
	if !ok {
		return nil, log.PushFatalStack("invalid subobjects array in configuration", nil, fields)
	}
	for i := range subobjectsConfig {
		path := Key(configPath, utils.NumToStr(i))
		fields := utils.AppendMapNew(fields, logger.Fields{"config_path": path})
		subobject := createSubobjectFnc()
		err := subobject.Init(cfg, log, vld, path)
		if err != nil {
			return nil, log.PushFatalStack("failed to initialize subobject", err, fields)
		}
		subobjects = append(subobjects, subobject)
	}

	return subobjects, nil
}

func LoadLogStringMapPlain[T any](cfg config.Config, log logger.Logger, configPath string, loggerFields ...logger.Fields) (map[string]T, error) {

	if !cfg.IsSet(configPath) {
		return nil, nil
	}

	fields := logger.AppendFieldsNew(logger.Fields{"config_path": configPath}, loggerFields...)
	m := make(map[string]T)

	mapSection := cfg.Get(configPath)
	mapConfig, ok := mapSection.(map[string]interface{})
	if !ok {
		return nil, log.PushFatalStack("invalid map section in configuration", nil, fields)
	}
	for key, value := range mapConfig {
		val, ok := value.(T)
		fullKey := Key(configPath, key)
		if !ok {
			fields["key"] = fullKey
			return nil, log.PushFatalStack("invalid value type in configuration", nil, fields)
		}
		m[key] = val

		logParameter(log, fullKey, fmt.Sprintf("%v", val), logger.InfoLevel)
	}

	return m, nil
}

func LoadLogStringMapInt(cfg config.Config, log logger.Logger, configPath string, loggerFields ...logger.Fields) (map[string]int, error) {
	return LoadLogStringMapPlain[int](cfg, log, configPath, loggerFields...)
}

func LoadLogStringMapString(cfg config.Config, log logger.Logger, configPath string, loggerFields ...logger.Fields) (map[string]string, error) {
	return LoadLogStringMapPlain[string](cfg, log, configPath, loggerFields...)
}
