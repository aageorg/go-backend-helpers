package pool

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type PoolStoreConfigI interface {
	GetPoolController() PoolController
}

type PoolStoreConfig struct {
	PoolController PoolController
}

func (p *PoolStoreConfig) GetPoolController() PoolController {
	return p.PoolController
}

type PoolStore interface {
	Pool(id string) (Pool, error)
	SelfPool() (Pool, error)
	PoolByName(name string) (Pool, error)
	Pools() []Pool
	PoolController() PoolController
}

type poolStoreConfig struct {
	POOL_NAME string
}

type PoolStoreBase struct {
	poolStoreConfig
	selfPool       Pool
	poolsByName    map[string]Pool
	poolsById      map[string]Pool
	poolController PoolController
}

func (p *PoolStoreBase) Config() interface{} {
	return &p.poolStoreConfig
}

func NewPoolStore(config ...PoolStoreConfigI) *PoolStoreBase {
	p := &PoolStoreBase{}
	p.poolsByName = make(map[string]Pool)
	p.poolsById = make(map[string]Pool)
	if len(config) != 0 {
		p.poolController = config[0].GetPoolController()
	}

	if p.poolController == nil {
		p.poolController = NewPoolController(&crud.DbCRUD{})
	}

	return p
}

func (p *PoolStoreBase) Init(ctx op_context.Context, configPath ...string) error {

	c := ctx.TraceInMethod("PoolStore.Init")
	defer ctx.TraceOutMethod()

	// load configuration
	err := object_config.LoadLogValidate(ctx.App().Cfg(), ctx.Logger(), ctx.App().Validator(), p, "pools", configPath...)
	if err != nil {
		msg := "failed to init PoolStore"
		c.SetMessage(msg)
		return ctx.Logger().PushFatalStack(msg, c.SetError(err))
	}

	loadServices := func(pool Pool) error {
		services, err := p.poolController.GetPoolBindings(ctx, pool.GetID())
		if err != nil {
			msg := "failed to load pool services"
			c.SetLoggerField("pool_name", pool.Name())
			c.SetLoggerField("pool_id", pool.GetID())
			c.SetMessage(msg)
			return ctx.Logger().PushFatalStack(msg, c.SetError(err))
		}
		pool.SetServices(services)
		return err
	}

	if p.POOL_NAME == "" {
		pools, _, err := p.poolController.GetPools(ctx, nil)
		if err != nil {
			msg := "failed to load pools"
			c.SetMessage(msg)
			return ctx.Logger().PushFatalStack(msg, c.SetError(err))
		}
		for _, pool := range pools {

			err = loadServices(pool)
			if err != nil {
				return err
			}

			p.poolsById[pool.GetID()] = pool
			p.poolsByName[pool.Name()] = pool
		}
	} else {
		pool, err := p.poolController.FindPool(ctx, p.POOL_NAME, true)
		if err != nil {
			msg := "failed to load self pool"
			c.SetMessage(msg)
			return ctx.Logger().PushFatalStack(msg, c.SetError(err))
		}
		if pool == nil {
			return c.SetErrorStr("self pool not found")
		}

		err = loadServices(pool)
		if err != nil {
			return err
		}

		p.poolsById[pool.GetID()] = pool
		p.poolsByName[pool.Name()] = pool
		p.selfPool = pool
	}

	// done
	return nil
}

func (p *PoolStoreBase) SelfPool() (Pool, error) {
	if p.selfPool == nil {
		return nil, errors.New("self pool undefined")
	}
	return p.selfPool, nil
}

func (p *PoolStoreBase) Pool(id string) (Pool, error) {
	pool, ok := p.poolsById[id]
	if !ok {
		return nil, errors.New("pool not found")
	}
	return pool, nil
}

func (p *PoolStoreBase) PoolByName(id string) (Pool, error) {
	pool, ok := p.poolsByName[id]
	if !ok {
		return nil, errors.New("pool not found")
	}
	return pool, nil
}

func (p *PoolStoreBase) Pools() []Pool {
	return utils.AllMapValues(p.poolsById)
}

func (p *PoolStoreBase) PoolController() PoolController {
	return p.poolController
}
