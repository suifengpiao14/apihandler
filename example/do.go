package example

import (
	"context"

	"github.com/suifengpiao14/apihandler"
)

type RequestID string

func Run(ctx context.Context, input string) (out string, err error) {
	key := RequestID("request_id")
	ctx = context.WithValue(ctx, key, "hello_world")
	method := "post"
	path := "/api/v1/adList"
	handler, err := apihandler.GetApi(method, path)
	if err != nil {
		return "", err
	}
	out, err = handler.Run(ctx, input)
	if err != nil {
		return "", err
	}
	return out, nil
}
