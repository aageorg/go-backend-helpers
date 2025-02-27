package pool_console

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

const ListPoolServicesCmd string = "list_pool_services"
const ListPoolServicesDescription string = "List pool services"

func ListPoolServices() Handler {
	a := &ListPoolServicesHandler{}
	a.Init(ListPoolServicesCmd, ListPoolServicesDescription)
	return a
}

type ListPoolServicesData struct {
	Name string `long:"name" description:"Short name of the pool" required:"true"`
}

type ListPoolServicesHandler struct {
	HandlerBase
	ListPoolServicesData
}

func (a *ListPoolServicesHandler) Data() interface{} {
	return &a.ListPoolServicesData
}

func (a *ListPoolServicesHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()
	services, err := controller.GetPoolBindings(ctx, a.Name, true)
	if err == nil {
		fmt.Printf("Services:\n\n%s\n\n", utils.DumpPrettyJson(services))
	}
	return err
}
