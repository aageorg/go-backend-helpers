package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const DisableServiceCmd string = "disable_service"
const DisableServiceDescription string = "Disable service"

func DisableService() poolsHandler {
	a := &DisableServiceHandler{}
	a.Init(DisableServiceCmd, DisableServiceDescription)
	return a
}

type DisableServiceHandler struct {
	poolsHandlerBase
	Service string `long:"pool" description:"Short name of the service" required:"true"`
}

func (a *DisableServiceHandler) Execute(args []string) error {

	ctx, controller := a.Context()
	defer ctx.Close()

	fields := db.Fields{}
	fields["active"] = false

	err := controller.UpdateService(ctx, a.Service, fields, true)
	if err == nil {
		pool, err := controller.FindPool(ctx, a.Service, true)
		if err == nil {
			if pool != nil {
				fmt.Printf("Updated service:\n\n%s\n\n", utils.DumpPrettyJson(pool))
			} else {
				fmt.Println("Pool not found")
			}
		}
	}
	return err
}
