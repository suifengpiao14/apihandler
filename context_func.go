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

type ApiType string
const (
	API_TYPE_QUERY ApiType = "query"
	API_TYPE_COMMAND ApiType = "command"
	_ApiTypeContentKey contextKey = "apiType"
)

//SetAPIType 记录 api 类型到上下文
func SetAPIType(api ApiInterface, apiType  ApiType) {
	ctx := api.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, _ApiTypeContentKey, apiType)
	api.SetContext(ctx)
}
//GetApiType 从上下文中获取 apiType
func GetApiType(apiInterface ApiInterface) (apiType ApiType) {
	value := apiInterface.GetContext().Value(_ApiTypeContentKey)
	apiType = value.(ApiType)
	return apiType
}

//ApiTypeIsQuery 判断是否为查询类型api
func ApiTypeIsQuery(apiType ApiType)bool{
	return apiType == API_TYPE_QUERY
}

//ApiTypeIsCommand 判断是否为命令类型api
func ApiTypeIsCommand(apiType ApiType)bool{
	return apiType == API_TYPE_COMMAND
}