package example

import (
	"context"
	"net/http"

	"github.com/suifengpiao14/apihandler"
	"github.com/suifengpiao14/logchan/v2"
	"github.com/suifengpiao14/stream"
)

func init() {
	logchan.SetLoggerWriter(stream.DefaultPrintStreamLog)
}

type RequestID string

func Run(ctx context.Context, input string) (out string, err error) {
	key := RequestID("request_id")
	ctx = context.WithValue(ctx, key, "hello_world")
	method := http.MethodPost
	path := "/api/v1/adList"
	api, err := apihandler.GetApi(ctx, method, path)
	if err != nil {
		return "", err
	}
	b, err := api.Run([]byte(input))
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}
