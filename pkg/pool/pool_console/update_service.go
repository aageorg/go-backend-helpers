package pool_console

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const UpdateServiceCmd string = "update_service"
const UpdateServiceDescription string = "Update service"

func UpdateService() Handler {
	a := &UpdateServiceHandler{}
	a.Init(UpdateServiceCmd, UpdateServiceDescription)
	return a
}

type UpdateServiceData struct {
	Service string `long:"service" description:"Short name of the service" required:"true"`
	Field   string `long:"field" description:"Field name" required:"true"`
	Value   string `long:"value" description:"Field value"`
}

type UpdateServiceHandler struct {
	HandlerBase
	UpdateServiceData
}

func (a *UpdateServiceHandler) Data() interface{} {
	return &a.UpdateServiceData
}

func (a *UpdateServiceHandler) Execute(args []string) error {

	ctx, controller, err := a.Context(a.Data())
	if err != nil {
		return err
	}
	defer ctx.Close()

	fields := db.Fields{}
	field := strings.ToLower(a.Field)
	if a.Field == "public_port" || a.Field == "private_port" {
		val, err := utils.StrToUint32(a.Value)
		if err != nil {
			fmt.Println("Value must be unsigned integer")
			return err
		}
		fields[field] = val
	} else {
		fields[field] = a.Value
	}
	vErr := validator.ValidateMap(ctx.App().Validator(), fields, &pool.PoolServiceBaseData{})
	if vErr != nil {
		app_context.ErrorLn("failed to validate fields")
		return vErr.Err
	}

	service, err := controller.UpdateService(ctx, a.Service, fields, true)
	if err != nil {
		return err
	}

	fmt.Printf("Updated service:\n\n%s\n\n", utils.DumpPrettyJson(service))
	return nil
}
