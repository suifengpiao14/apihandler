package example

import (
	"context"

	"github.com/suifengpiao14/controllerhandler"
)

type RequestID string

func NewAdListInput() (adListInput *AdListInput) {
	adListInput = &AdListInput{}
	adListInput.SetDoFn(AdListDoFn)
	return adListInput
}

func Run(ctx context.Context, input string) (out string, err error) {
	key := RequestID("request_id")
	ctx = context.WithValue(ctx, key, "hello_world")
	adListInput := NewAdListInput()
	handler, err := controllerhandler.NewApi(adListInput)
	if err != nil {
		return "", err
	}
	out, err = handler.Run(ctx, input)
	if err != nil {
		return "", err
	}
	return out, nil
}
