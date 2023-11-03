package apihandler

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type contextKey string

type httpRequestAndResponseWriter struct {
	req *http.Request
	w   http.ResponseWriter
}

var (
	httpRequestAndResponseWriterKey contextKey = "httpRequestAndResponseWriter"
	CONTEXT_NOT_FOUND_KEY                      = errors.New("not found key")
	CONTEXT_NOT_EXCEPT                         = errors.New("not except type")
)

//SetHttpRequestAndResponseWriter 记录http 请求上下文
func SetHttpRequestAndResponseWriter(api ApiInterface, req *http.Request, w http.ResponseWriter) {
	ctx := api.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	value := &httpRequestAndResponseWriter{
		req: req,
		w:   w,
	}
	ctx = context.WithValue(ctx, httpRequestAndResponseWriterKey, value)
	api.SetContext(ctx)
}

func GetHttpRequestAndResponseWriter(apiInterface ApiInterface) (req *http.Request, w http.ResponseWriter, err error) {
	value := apiInterface.GetContext().Value(httpRequestAndResponseWriterKey)
	if value == nil {
		err = errors.WithMessagef(CONTEXT_NOT_FOUND_KEY, "key:%s", httpRequestAndResponseWriterKey)
		return nil, nil, err
	}
	_httpRequestAndResponseWriter, ok := value.(*httpRequestAndResponseWriter)
	if !ok {
		err = errors.WithMessagef(CONTEXT_NOT_EXCEPT, "except:*httpRequestAndResponseWriter,got:%T", value)
		return nil, nil, err
	}
	return _httpRequestAndResponseWriter.req, _httpRequestAndResponseWriter.w, nil
}
