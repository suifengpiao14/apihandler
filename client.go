package apihandler

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/gojsonschemavalidator"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/logchan/v2"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
)

type HttpRequestFunc func(ctx context.Context, client ClientInterface, w http.ResponseWriter, r *http.Request) (err error) // 此处只返回error,确保输出写入到w

type ClientInterface interface {
	GetRequestHandlerFunc() (httpRequestFunc HttpRequestFunc)
	GetDoFn() (doFn func(ctx context.Context) (out OutputI, err error))
	GetInputSchema() (lineschema string)
	GetOutputSchema() (lineschema string)
	GetRoute() (method string, path string)
	Init()
	GetDescription() (title string, description string)
	GetName() (domain string, name string)
}

type DefaultImplementClientFuncs struct{}

func (e *DefaultImplementClientFuncs) GetInputSchema() (lineschema string) {
	return ""
}
func (e *DefaultImplementClientFuncs) GetOutputSchema() (lineschema string) {
	return ""
}

func (e *DefaultImplementClientFuncs) Init() {
}

func (e *DefaultImplementClientFuncs) GetRequestHandlerFunc() (httpRequestFunc HttpRequestFunc) {
	return DefaultHttpRequestFunc
}

type LogInfoClientRun struct {
	Context        context.Context
	Input          string
	DefaultJson    string
	MergedDefault  string
	Err            error `json:"error"`
	FormattedInput string
	OriginalOut    string
	Out            string
	logchan.EmptyLogInfo
}

func (l *LogInfoClientRun) GetName() logchan.LogName {
	return LOG_INFO_EXEC_Client_HANDLER
}
func (l *LogInfoClientRun) Error() error {
	return l.Err
}

const (
	LOG_INFO_EXEC_Client_HANDLER LogName = "LogInfoExecClientHandler"
)

type _Client struct {
	ClientInterface
	inputFormatGjsonPath  string
	defaultJson           string
	outputFormatGjsonPath string
	validateInputLoader   gojsonschema.JSONLoader
	validateOutputLoader  gojsonschema.JSONLoader
}

var clientMap sync.Map

const (
	clientMap_route_add_key = "___all_api_add___"
)

// RegisterClient 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func RegisterClient(ClientInterface ClientInterface) (err error) {
	method, path := ClientInterface.GetRoute()
	key := getRouteKey(method, path)
	// 以下初始化可以复用,线程安全
	api := &_Client{
		ClientInterface: ClientInterface,
	}
	inputSchema := ClientInterface.GetInputSchema()
	if inputSchema != "" {
		api.validateInputLoader, err = newJsonschemaLoader(inputSchema)
		if err != nil {
			return err
		}
		inputLineSchema, err := jsonschemaline.ParseJsonschemaline(inputSchema)
		if err != nil {
			return err
		}
		api.inputFormatGjsonPath = inputLineSchema.GjsonPathWithDefaultFormat(true)
		defaultInputJson, err := inputLineSchema.DefaultJson()
		if err != nil {
			err = errors.WithMessage(err, "get input default json error")
			return err
		}
		api.defaultJson = defaultInputJson.Json
	}
	outputSchema := ClientInterface.GetOutputSchema()
	if outputSchema != "" {
		api.validateOutputLoader, err = newJsonschemaLoader(outputSchema)
		if err != nil {
			return err
		}
		outputLineSchema, err := jsonschemaline.ParseJsonschemaline(outputSchema)
		if err != nil {
			return err
		}
		api.outputFormatGjsonPath = outputLineSchema.GjsonPathWithDefaultFormat(true)
	}
	clientMap.Store(key, api)
	routes := make(map[string][2]string, 0)
	if routesI, ok := clientMap.Load(clientMap_route_add_key); ok {
		if old, ok := routesI.(map[string][2]string); ok {
			routes = old
		}
	}
	route := [2]string{method, path}
	routes[key] = route
	clientMap.Store(clientMap_route_add_key, routes)
	return nil
}

func GetClient(method string, path string) (api _Client, err error) {
	key := getRouteKey(method, path)
	apiAny, ok := clientMap.Load(key)
	if !ok {
		return api, errors.WithMessagef(API_NOT_FOUND, "method:%s,path:%s", method, path)
	}
	exitsApi := apiAny.(*_Client)
	rt := reflect.TypeOf(exitsApi.ClientInterface).Elem()
	rv := reflect.New(rt)
	ClientInterface := rv.Interface().(ClientInterface)
	ClientInterface.Init()
	api = _Client{
		ClientInterface:       ClientInterface,
		validateInputLoader:   exitsApi.validateInputLoader,
		validateOutputLoader:  exitsApi.validateOutputLoader,
		inputFormatGjsonPath:  exitsApi.inputFormatGjsonPath,
		outputFormatGjsonPath: exitsApi.outputFormatGjsonPath,
		defaultJson:           exitsApi.defaultJson,
	}
	return api, nil
}

func (a _Client) inputValidate(input string) (err error) {
	if a.validateInputLoader == nil {
		return nil
	}
	inputStr := string(input)
	err = gojsonschemavalidator.Validate(inputStr, a.validateInputLoader)
	if err != nil {
		return err
	}
	return nil
}
func (a _Client) outputValidate(output string) (err error) {
	outputStr := string(output)
	if a.validateOutputLoader == nil {
		return nil
	}
	err = gojsonschemavalidator.Validate(outputStr, a.validateOutputLoader)
	if err != nil {
		return err
	}
	return nil
}

func (a _Client) modifyTypeByFormat(input string, formatGjsonPath string) (formattedInput string, err error) {
	formattedInput = input
	if formatGjsonPath == "" {
		return formattedInput, nil
	}
	formattedInput = gjson.Get(input, formatGjsonPath).String()
	return formattedInput, nil
}

func (a _Client) convertInput(input string) (err error) {
	err = json.Unmarshal([]byte(input), a.ClientInterface)
	if err != nil {
		return err
	}
	return nil
}

func (a _Client) RunRequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	httpRequestFunc := a.ClientInterface.GetRequestHandlerFunc()
	if httpRequestFunc == nil {
		err = errors.Errorf("GetHttpHandlerFunc return nil: %v", a)
		return err
	}
	err = FillterAuth(w, r) //这个中间件，书写方式后续可以优化
	if err != nil {
		return err
	}
	err = httpRequestFunc(ctx, a, w, r)
	return err
}

func (a _Client) Run(ctx context.Context, input string) (out string, err error) {
	logInfo := LogInfoClientRun{
		Context:     ctx,
		Input:       input,
		DefaultJson: a.defaultJson,
	}
	defer func() {
		logchan.SendLogInfo(&logInfo)
	}()

	if a.ClientInterface == nil {
		err = errors.Errorf("handlerInterface required %v", a)
		return "", err
	}
	if a.ClientInterface.GetDoFn() == nil { //此处只先判断,不取值,等后续将input值填充后再获取
		err = errors.Errorf("doFn required %v", a.ClientInterface)
		return "", err
	}

	// 合并默认值
	if a.defaultJson != "" {
		input, err = jsonschemaline.MergeDefault(input, a.defaultJson)
		if err != nil {
			err = errors.WithMessage(err, "merge default value error")
			return "", err
		}
		logInfo.MergedDefault = input
	}
	err = a.inputValidate(input)
	if err != nil {
		return "", err
	}
	//将format 中 int,float,bool 应用到数据
	formattedInput, err := a.modifyTypeByFormat(input, a.inputFormatGjsonPath)
	if err != nil {
		return "", err
	}
	logInfo.FormattedInput = formattedInput
	err = a.convertInput(formattedInput)
	if err != nil {
		return "", err
	}
	doFn := a.ClientInterface.GetDoFn()
	outI, err := doFn(ctx)
	if err != nil {
		return "", err
	}
	if funcs.IsNil(outI) {
		err = errors.New("response not be nil ")
		err = errors.WithMessage(err, "github.com/suifengpiao14/apihandler._Client.Run")
		return "", err
	}
	originalOut, err := outI.String()
	if err != nil {
		return "", err
	}
	logInfo.OriginalOut = originalOut
	out, err = a.modifyTypeByFormat(originalOut, a.outputFormatGjsonPath)
	if err != nil {
		return "", err
	}
	logInfo.Out = out
	err = a.outputValidate(out)
	if err != nil {
		return "", err
	}
	return out, nil
}

func DefaultHttpRequestFunc(ctx context.Context, client ClientInterface, w http.ResponseWriter, r *http.Request) (err error) {
	return
}
