package user_manager

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type UserManager interface {
	MakeAuthUser() auth.User
	Find(ctx op_context.Context, fields db.Fields, user interface{}) (bool, error)
	Create(ctx op_context.Context, user common.Object) error
	Update(ctx op_context.Context, user common.Object, fields db.Fields) error

	GetUsers(ctx op_context.Context, filter *db.Filter, users interface{}) error

	ValidateLogin(login string) error
}

type WithUserManager interface {
	UserManager() UserManager
}

func FindByLogin(manager UserManager, ctx op_context.Context, login string, user interface{}) (bool, error) {
	return manager.Find(ctx, db.Fields{"login": login}, user)
}

type UserManagerBase struct {
}

func (m *UserManagerBase) Find(ctx op_context.Context, fields db.Fields, user interface{}) (bool, error) {

	c := ctx.TraceInMethod("UserManagerBase.Find")
	defer ctx.TraceOutMethod()

	found, err := op_context.DB(ctx).FindByFields(ctx, fields, user)
	if err != nil {
		return found, c.SetError(err)
	}

	return found, nil
}

func (m *UserManagerBase) Create(ctx op_context.Context, user common.Object) error {
	c := ctx.TraceInMethod("UserManagerBase.Create")
	defer ctx.TraceOutMethod()

	err := op_context.DB(ctx).Create(ctx, user)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (m *UserManagerBase) Update(ctx op_context.Context, user common.Object, fields db.Fields) error {
	c := ctx.TraceInMethod("UserManagerBase.Update")
	defer ctx.TraceOutMethod()

	err := db.Update(op_context.DB(ctx), ctx, user, fields)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (m *UserManagerBase) GetUsers(ctx op_context.Context, filter *db.Filter, users interface{}) error {
	return op_context.LoadObjects(ctx, "", filter, users)
}
