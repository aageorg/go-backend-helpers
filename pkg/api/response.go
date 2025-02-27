package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

const (
	TargetSelf   = "self"
	TargetParent = "parent"
	TargetChild  = "parent"
)

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func ResponseGenericError(responseError *ResponseError) generic_error.Error {
	if responseError == nil {
		return nil
	}
	e := generic_error.New(responseError.Code, responseError.Message)
	e.SetDetails(responseError.Details)
	return e
}

type Response interface {
	WithHateoasLinks
}

type ResponseStub struct {
	HateoasLinksStub
}

type ResponseBase struct {
	HateoasLinksContainer
}

type ResponseCount struct {
	Count int64 `json:"count,omitempty"`
}

type ResponseExists struct {
	ResponseStub
	Exists bool `json:"exists"`
}

type ResponseListI interface {
	Response
	ItemCount() int
	ItemId(index int) string
	MakeItemLinks()
	SetItemLink(index int, link *HateoasLinksItem)
}

type ResponseList[T common.WithID] struct {
	ResponseCount
	Items []T `json:"items"`

	ResponseBase
	ItemLinks []*HateoasLinksItem `json:"_item_links,omitempty"`
}

func (r *ResponseList[T]) ItemCount() int {
	return len(r.Items)
}

func (r *ResponseList[T]) ItemId(index int) string {
	return r.Items[index].GetID()
}

func (r *ResponseList[T]) MakeItemLinks() {
	r.ItemLinks = make([]*HateoasLinksItem, r.ItemCount())
}

func (r *ResponseList[T]) SetItemLink(index int, link *HateoasLinksItem) {
	r.ItemLinks[index] = link
}
