package tenancy_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_manager"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func (t *TenancyClient) SetPath(ctx op_context.Context, id string, path string, idIsDisplay ...bool) error {

	// setup
	c := ctx.TraceInMethod("TenancyClient.SetPath")
	defer ctx.TraceOutMethod()

	// setup ID
	tenancyId, _, err := tenancy_manager.TenancyId(t, ctx, id, idIsDisplay...)
	if err != nil {
		c.SetMessage("failed to get ID")
		return c.SetError(err)
	}

	// create command
	handler := api_client.NewHandlerCmd(&multitenancy.WithPath{PATH: path})

	// prepare and exec handler
	op := api.OperationAsResource(t.TenancyResource, "path", tenancyId, tenancy_api.SetPath())
	err = op.Exec(ctx, api_client.MakeOperationHandler(t.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil

}
