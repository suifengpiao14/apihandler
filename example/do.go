package example

import (
	"context"
	"net/http"

	"github.com/suifengpiao14/apihandler"
)

type RequestID string

func Run(ctx context.Context, input string) (out string, err error) {
	key := RequestID("request_id")
	ctx = context.WithValue(ctx, key, "hello_world")
	method := http.MethodPost
	path := "/api/v1/adList"
	handler, err := apihandler.GetApi(ctx, method, path)
	if err != nil {
		return "", err
	}
	b, err := apihandler.Run(handler, []byte(input))
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}
