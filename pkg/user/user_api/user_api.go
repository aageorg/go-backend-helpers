package user_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

func PrepareResources(userTypeName ...string) (userType string, serviceName string, collectionResource api.Resource, userResource api.Resource) {

	userType = utils.OptionalArg("user", userTypeName...)
	serviceName = utils.ConcatStrings(userType, "s")

	userResource = UserResource(userType)
	collectionResource = userResource.Parent()

	return
}

func NamedUserResource(id string, userTypeName ...string) (userResource api.Resource) {
	r := UserResource(userTypeName...)
	r.SetId(id)
	return r
}

func UserResource(resourceType ...string) api.Resource {
	return api.NamedResource(utils.OptionalArg("user", resourceType...))
}

type UserResponse[T user.User] struct {
	api.ResponseHateous
	User T `json:"user"`
}

type ListResponse[T any] struct {
	api.ResponseCount
	api.ResponseHateous
	Users *[]T `json:"users"`
}

func List() api.Operation {
	return api.NewOperation("list", access_control.Read)
}

func Add(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("add_", name), access_control.Create)
}

func Find(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("find_", name), access_control.Get)
}

func SetPassword(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("find_", name, "_password"), access_control.Put)
}

func SetEmail(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("find_", name, "_email"), access_control.Put)
}

func SetPhone(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("find_", name, "_phone"), access_control.Put)
}

func SetBlocked(name string) api.Operation {
	return api.NewOperation(utils.ConcatStrings("find_", name, "_blocked"), access_control.Put)
}
