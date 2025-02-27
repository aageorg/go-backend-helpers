package multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func UpgradeTenancyDatabase(ctx op_context.Context, tenancy Tenancy, dbModels *TenancyDbModels) error {

	// setup
	loggerFields := logger.Fields{"tenancy": tenancy.GetID(), "tenancy_db": tenancy.DbName()}
	var err error
	c := ctx.TraceInMethod("multitenancy.UpgradeTenancyDatabase", loggerFields)
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// migrate internal implicit models
	err = tenancy.Db().AutoMigrate(ctx, DbInternalModels())
	if err != nil {
		c.SetMessage("failed to upgrade internal models in tenancy database")
		return err
	}

	// migrate explicit models
	err = tenancy.Db().AutoMigrate(ctx, dbModels.DbModels)
	if err != nil {
		c.SetMessage("failed to upgrade ordinary models tenancy database")
		return err
	}

	// migrate partitioned db models
	err = tenancy.Db().PartitionedMonthAutoMigrate(ctx, dbModels.PartitionedDbModels)
	if err != nil {
		c.SetMessage("failed to upgrade partitioned models in tenancy database")
		return err
	}

	// done
	return nil
}

func UpgradeTenancyDatabases(ctx op_context.Context, multitenancy Multitenancy, dbModels *TenancyDbModels) error {

	c := ctx.TraceInMethod("multitenancy.UpgradeTenancyDatabases")
	defer ctx.TraceOutMethod()

	for _, tenancy := range multitenancy.Tenancies() {
		err := UpgradeTenancyDatabase(ctx, tenancy, dbModels)
		if err != nil {
			return c.SetError(err)
		}
	}

	return nil
}
