package api_server

import (
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

// Interface of response of server API.
type Response interface {
	Message() interface{}
	SetMessage(message api.Response)
	Request() Request
	SetRequest(request Request)

	Text() string
	SetText(text string)

	SetRedirectPath(path string)
	RedirectPath() string
}

type ResponseBase struct {
	message              interface{}
	request              Request
	text                 string
	redirectResourcePath string
}

func (r *ResponseBase) Message() interface{} {
	return r.message
}

func (r *ResponseBase) SetMessage(message api.Response) {
	if r.request.Server().IsHateoas() {
		api.InjectHateoasLinksToObject(r.request.Endpoint().Resource(), message)
	}
	r.message = message
}

func SetResponseList(r Request, response api.ResponseListI, resourceType ...string) {

	if r.Server().IsHateoas() {
		resource := r.Endpoint().Resource()
		rType := utils.OptionalArg(resource.Type(), resourceType...)
		api.HateoasList(response, resource, rType)
	}

	r.Response().SetMessage(response)
}

func (r *ResponseBase) SetRequest(request Request) {
	r.request = request
}

func (r *ResponseBase) Request() Request {
	return r.request
}

func (r *ResponseBase) SetText(text string) {
	r.text = text
}

func (r *ResponseBase) Text() string {
	return r.text
}

func (r *ResponseBase) SetRedirectPath(path string) {
	r.redirectResourcePath = path
}

func (r *ResponseBase) RedirectPath() string {
	return r.redirectResourcePath
}
