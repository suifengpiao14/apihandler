package example

import (
	"context"

	"github.com/suifengpiao14/controllerhandler"
)

func NewAdListHandler() (handler *controllerhandler.Handler, err error) {
	adListInput := &AdListInput{
		Output: AdListOutput{},
	}
	handler = controllerhandler.NewHandler(adListInput)
	return handler, nil
}

func Run(ctx context.Context, input string) (out string, err error) {
	adListHandler, err := NewAdListHandler()
	if err != nil {
		return "", err
	}
	out, err = adListHandler.Run(ctx, input)
	if err != nil {
		return "", err
	}
	return out, nil
}
