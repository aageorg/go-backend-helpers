package tenancy_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/tenancy_api"
)

type SetPathEndpoint struct {
	TenancyUpdateEndpoint
}

func (s *SetPathEndpoint) HandleRequest(request api_server.Request) error {

	// setup
	c := request.TraceInMethod("tenancy.SetPath")
	defer request.TraceOutMethod()

	// parse command
	cmd := &multitenancy.WithPath{}
	err := request.ParseValidate(cmd)
	if err != nil {
		c.SetMessage("failed to parse/validate command")
		return err
	}

	// apply
	err = s.service.Tenancies.SetPath(request, request.GetResourceId(tenancy_api.TenancyResource), cmd.Path())
	if err != nil {
		return c.SetError(err)
	}

	// done
	return nil
}

func SetPath(s *TenancyService) *SetPathEndpoint {
	e := &SetPathEndpoint{}
	e.Construct(s, e, "path", tenancy_api.SetPath())
	return e
}
