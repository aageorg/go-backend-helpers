package user_client

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_api"
)

type SetEmail = SetterHandler[user_api.SetEmailCmd]

func (u *UserClient[U]) SetEmail(ctx op_context.Context, id string, email string, idIsLogin ...bool) error {

	// setup
	c := ctx.TraceInMethod("UserClient.SetEmail")
	defer ctx.TraceOutMethod()

	// if idIsLogin then first find user
	userId, err := u.GetUserId(ctx, id, idIsLogin...)
	if err != nil {
		c.SetMessage("failed to get user ID")
		return c.SetError(err)
	}

	// create command
	handler := &SetEmail{}
	handler.Cmd.Email = email

	// prepare and exec handler
	err = u.UserOperation(userId, "email", user_api.SetEmail()).Exec(ctx, api_client.MakeOperationHandler(u.Client(), handler))
	if err != nil {
		c.SetMessage("failed to exec operation")
		return c.SetError(err)
	}

	// done
	return nil
}
