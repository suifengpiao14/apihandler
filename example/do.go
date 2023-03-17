package example

import (
	"context"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/apihandler"
)

func init() {
	apihandler.RegisterApi(&AdListInput{})
}

type RequestID string

func Run(ctx context.Context, input string) (out string, err error) {
	key := RequestID("request_id")
	ctx = context.WithValue(ctx, key, "hello_world")
	method := "post"
	path := "/api/v1/adList"
	handler, ok := apihandler.GetApi(method, path)
	if !ok {
		err = errors.Errorf("not found hanndler by method:%s,path:%s", method, path)
		return "", err
	}
	out, err = handler.Run(ctx, input)
	if err != nil {
		return "", err
	}
	return out, nil
}
