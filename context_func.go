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
	capiKey                         contextKey = "_CApi"
	httpRequestAndResponseWriterKey contextKey = "httpRequestAndResponseWriter"
	logInfoApiRunKey                contextKey = "logInfoApiRun"
	CONTEXT_NOT_FOUND_KEY                      = errors.New("not found key")
	CONTEXT_NOT_EXCEPT                         = errors.New("not except")
)

//setCAPI 从 ApiInterface 中获取编译好的_CApi 包内使用
func setCAPI(apiInterface ApiInterface, capi *_CApi) {
	ctx := apiInterface.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, capiKey, capi)
	apiInterface.SetContext(ctx)
}

//GetCAPI 从 ApiInterface 中获取编译好的_CApi
func GetCAPI(apiInterface ApiInterface) (_capi *_CApi, err error) {
	value := apiInterface.GetContext().Value(capiKey)
	if value == nil {
		err = errors.WithMessagef(CONTEXT_NOT_FOUND_KEY, "key:%s", capiKey)
		return nil, err
	}
	_capi, ok := value.(*_CApi)
	if !ok {
		err = errors.WithMessagef(CONTEXT_NOT_EXCEPT, "except:*_CApi,got:%T", value)
		return nil, err
	}
	return _capi, nil
}

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

//setLogInfoApiRun 记录api执行上下文，设置到ctx中，方便在doFn 内部获取使用,仅供内部设置
func setLogInfoApiRun(api ApiInterface, loginInfo *LogInfoApiRun) {
	ctx := api.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, logInfoApiRunKey, loginInfo)
	api.SetContext(ctx)
}

//GetLogInfoApiRun 从上下文中获取日志记录对象
func GetLogInfoApiRun(apiInterface ApiInterface) (loginInfo *LogInfoApiRun, err error) {
	value := apiInterface.GetContext().Value(logInfoApiRunKey)
	if value == nil {
		err = errors.WithMessagef(CONTEXT_NOT_FOUND_KEY, "key:%s", logInfoApiRunKey)
		return nil, err
	}
	loginInfo, ok := value.(*LogInfoApiRun)
	if !ok {
		err = errors.WithMessagef(CONTEXT_NOT_EXCEPT, "except:*LogInfoApiRun,got:%T", value)
		return nil, err
	}
	return loginInfo, nil
}
