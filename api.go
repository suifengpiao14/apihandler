package apihandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/suifengpiao14/apihandler/auth"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/logchan/v2"
	"github.com/suifengpiao14/stream"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
)

func init() {
	// 注册默认鉴权服务
	auth.RegisterAuthFunc(auth.CasDoorAuthFunc)
}

var API_NOT_FOUND = errors.Errorf("not found")
var (
	ERROR_NOT_IMPLEMENTED = errors.New("not implemented")
)

type ApiInterface interface {
	GetDoFn() (doFn func(ctx context.Context) (out OutputI, err error))
	GetInputSchema() (lineschema string)
	GetOutputSchema() (lineschema string)
	GetRoute() (method string, path string)
	Init()
	GetDescription() (title string, description string)
	GetName() (domain string, name string)
	GetConfig() (cfg ApiConfig)
	SetContext(ctx context.Context)
	GetContext() (ctx context.Context)
}

type ApiConfig struct {
	Auth     bool          `json:"auth"`     // 需要鉴权
	Throttle time.Duration `json:"throttle"` // 节流,一定时间内只执行一次,防止多次连续点击
}

type LogInfoApiRun struct {
	Input          string
	DefaultJson    string
	MergedDefault  string
	Err            error `json:"error"`
	FormattedInput string
	OriginalOut    string
	Out            string
	More           interface{}
	logchan.EmptyLogInfo
}

func (l *LogInfoApiRun) GetName() logchan.LogName {
	return LOG_INFO_EXEC_API_HANDLER
}
func (l *LogInfoApiRun) Error() error {
	return l.Err
}

// DefaultPrintLogInfoApiRun 默认api执行日志打印函数
func DefaultPrintLogInfoApiRun(logInfo logchan.LogInforInterface, typeName logchan.LogName, err error) {
	if typeName != LOG_INFO_EXEC_API_HANDLER {
		return
	}
	apiRunLogInfo, ok := logInfo.(*LogInfoApiRun)
	if !ok {
		return
	}
	if err != nil {
		_, err1 := fmt.Fprintf(logchan.LogWriter, "%s|loginInfo:%s|\nerror:%s\n|input:%s\n", logchan.DefaultPrintLog(apiRunLogInfo), apiRunLogInfo.GetName(), err.Error(), apiRunLogInfo.Input)
		if err1 != nil {
			fmt.Printf("err: DefaultPrintLogInfoApiRun fmt.Fprintf:%s\n", err1.Error())
		}
		return
	}
	moreb, _ := json.Marshal(apiRunLogInfo.More)
	more := string(moreb)
	_, err1 := fmt.Fprintf(logchan.LogWriter, "%s|input:%s|output:%s|more:%s\n", logchan.DefaultPrintLog(apiRunLogInfo), apiRunLogInfo.Input, apiRunLogInfo.Out, more)
	if err1 != nil {
		fmt.Printf("err: DefaultPrintLogInfoApiRun fmt.Fprintf:%s\n", err1.Error())
	}
}

type LogName string

func (logName LogName) String() (name string) {
	return string(logName)
}

const (
	LOG_INFO_EXEC_API_HANDLER LogName = "LogInfoExecApiHandler"
)

// DefaultImplementFuncs 可选部分接口函数
type DefaultImplementFuncs struct {
	ctx context.Context
}

func (e *DefaultImplementFuncs) GetInputSchema() (lineschema string) {
	return ""
}
func (e *DefaultImplementFuncs) GetOutputSchema() (lineschema string) {
	return ""
}

func (e *DefaultImplementFuncs) Init() {
}

func (e *DefaultImplementFuncs) GetConfig() (cfg ApiConfig) {
	return ApiConfig{
		Auth: true,
	}
}

func (e *DefaultImplementFuncs) SetContext(ctx context.Context) {
	e.ctx = ctx

}
func (e *DefaultImplementFuncs) GetContext() (ctx context.Context) {
	return e.ctx
}

type OutputI interface {
	String() (out string, err error)
}

type OutputString string

func (output *OutputString) String() (out string, err error) {
	out = string(*output)
	return out, nil
}

type _OutputJson struct {
	v any
}

func (output _OutputJson) String() (out string, err error) {
	b, err := json.Marshal(output.v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func OutputJson(v any) OutputI {
	return _OutputJson{
		v: v,
	}
}

func getRouteKey(method string, path string) (key string) {
	return fmt.Sprintf("%s_%s", strings.ToLower(method), path)
}

func JsonMarshal(o interface{}) (out string, err error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}

type _CApi struct {
	ApiInterface
	inputFormatGjsonPath  string
	defaultJson           string
	outputFormatGjsonPath string
	validateInputLoader   gojsonschema.JSONLoader
	validateOutputLoader  gojsonschema.JSONLoader
}

var apiMap sync.Map

const (
	apiMap_route_add_key = "___all_route_add___"
)

// RegisterApi 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func RegisterApi(apiInterface ApiInterface) (err error) {
	method, path := apiInterface.GetRoute()
	key := getRouteKey(method, path)
	// 以下初始化可以复用,线程安全
	api := &_CApi{
		ApiInterface: apiInterface,
	}
	inputSchema := apiInterface.GetInputSchema()
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
	outputSchema := apiInterface.GetOutputSchema()
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
	apiMap.Store(key, api)
	routes := make(map[string][2]string, 0)
	if routesI, ok := apiMap.Load(apiMap_route_add_key); ok {
		if old, ok := routesI.(map[string][2]string); ok {
			routes = old
		}
	}
	route := [2]string{method, path}
	routes[key] = route
	apiMap.Store(apiMap_route_add_key, routes)
	return nil
}

type APIProfile struct {
	Domain      string `json:"domain"`      // 领域
	Name        string `json:"name"`        // 名称 唯一键
	Title       string `json:"title"`       // 标题
	Method      string `json:"method"`      // 请求方法
	Path        string `json:"path"`        // 请求路径
	Description string `json:"description"` //描述
}

func GetAPIProfile(api ApiInterface) (apiProfile APIProfile) {
	domain, name := api.GetName()
	title, description := api.GetDescription()
	method, path := api.GetRoute()
	apiProfile = APIProfile{
		Domain:      domain,      // 领域
		Name:        name,        // 名称 唯一键
		Title:       title,       // 标题
		Method:      method,      // 请求方法
		Path:        path,        // 请求路径
		Description: description, //描述
	}
	return apiProfile
}

func GetAllAPIProfile() (apiProfiles []APIProfile, err error) {
	apiProfiles = make([]APIProfile, 0)
	apis, err := getAllAPI()
	if err != nil {
		return nil, err
	}
	for _, api := range apis {
		apiProfile := GetAPIProfile(api)
		validate := validator.New()
		err = validate.Struct(apiProfile)
		if err != nil {
			method, path := api.GetRoute()
			err = errors.WithMessagef(err, "method:%s,path:%s", method, path)
			return nil, err
		}
		apiProfiles = append(apiProfiles, apiProfile)
	}
	return apiProfiles, nil
}

func RegisterRouteFn(routeFn func(method string, path string)) {
	routes := GetAllRoute()
	for _, route := range routes {
		method, path := route[0], route[1]
		routeFn(method, path)
	}
}

func getAllAPI() (apis []ApiInterface, err error) {
	routes := GetAllRoute()
	apis = make([]ApiInterface, 0)
	for _, route := range routes {
		method, path := route[0], route[1]
		api, err := GetApi(context.Background(), method, path)
		if err != nil {
			return nil, err
		}
		apis = append(apis, api)
	}
	return apis, nil
}

// GetAllRoute 获取已注册的所有api route
func GetAllRoute() (routes [][2]string) {
	routes = make([][2]string, 0)
	if routesI, ok := apiMap.Load(apiMap_route_add_key); ok {
		if tmp, ok := routesI.(map[string][2]string); ok {
			for _, route := range tmp {
				routes = append(routes, route)
			}
		}
	}

	return routes
}

// Run 启动运行
func Run(ctx context.Context, method string, path string, input string) (out string, err error) {
	api, err := GetApi(ctx, method, path)
	if err != nil {
		return "", err
	}
	out, err = api.Run(ctx, string(input))
	if err != nil {
		return "", err
	}
	return out, nil

}

func GetApi(ctx context.Context, method string, path string) (api _CApi, err error) {
	key := getRouteKey(method, path)
	apiAny, ok := apiMap.Load(key)
	if !ok {
		return api, errors.WithMessagef(API_NOT_FOUND, "method:%s,path:%s", method, path)
	}
	exitsApi := apiAny.(*_CApi)
	rt := reflect.TypeOf(exitsApi.ApiInterface).Elem()
	rv := reflect.New(rt)
	apiInterface := rv.Interface().(ApiInterface)
	apiInterface.Init()
	api = _CApi{
		ApiInterface:          apiInterface,
		validateInputLoader:   exitsApi.validateInputLoader,
		validateOutputLoader:  exitsApi.validateOutputLoader,
		inputFormatGjsonPath:  exitsApi.inputFormatGjsonPath,
		outputFormatGjsonPath: exitsApi.outputFormatGjsonPath,
		defaultJson:           exitsApi.defaultJson,
	}
	api.initContext(ctx)
	return api, nil
}

func (a _CApi) initContext(ctx context.Context) {
	a.ApiInterface.SetContext(ctx)
	setCAPI(a.ApiInterface, &a)
}

func (a _CApi) Run(ctx context.Context, input string) (out string, err error) {

	if a.ApiInterface == nil {
		err = errors.Errorf("handlerInterface required %v", a)
		return "", err
	}
	dostream := stream.NewStream(
		ErrorHandlerFn(),
		MakeUnPackHandler(),
		MakeMergeDefaultHandler([]byte(a.defaultJson)),
		MakeValidateHandler(a.validateInputLoader),
		MakeFormatHandler(a.inputFormatGjsonPath),
		MakeUnmarshalHandler(a.ApiInterface),
		func(ctx context.Context, input []byte) (out []byte, err error) {
			doFn := a.ApiInterface.GetDoFn()
			outI, err := doFn(ctx)
			if err != nil {
				return nil, err
			}
			if funcs.IsNil(outI) {
				err = errors.New("response not be nil ")
				err = errors.WithMessage(err, "github.com/suifengpiao14/apihandler._CApi.Run")
				return nil, err
			}
			originalOut, err := outI.String()
			if err != nil {
				return nil, err
			}
			return []byte(originalOut), nil
		},
		MakeFormatHandler(a.outputFormatGjsonPath),
		MakeValidateHandler(a.validateOutputLoader),
		MakePackHandler(),
	)
	if dostream == nil {
		err = errors.Errorf("work stream required %v", a.ApiInterface)
		return "", err
	}
	outB, err := dostream.Go(a.GetContext(), []byte(input))
	if err != nil {
		return "", err
	}
	return string(outB), nil
}

func newJsonschemaLoader(lineSchemaStr string) (jsonschemaLoader gojsonschema.JSONLoader, err error) {
	if lineSchemaStr == "" {
		err = errors.Errorf("NewJsonschemaLoader: arg lineSchemaStr required,got empty")
		return nil, err
	}
	inputlineSchema, err := jsonschemaline.ParseJsonschemaline(lineSchemaStr)
	if err != nil {
		return nil, err
	}
	jsb, err := inputlineSchema.JsonSchema()
	if err != nil {
		return nil, err
	}
	jsonschemaStr := string(jsb)
	jsonschemaLoader = gojsonschema.NewStringLoader(jsonschemaStr)
	return jsonschemaLoader, nil
}

var (
	Error_Content_Type_Required = errors.New("http request header Content-Type required")
)

// RequestInputToJson 统一获取 query,header,body 参数
func RequestInputToJson(r *http.Request, useArrInQueryAndHead bool) (reqInput []byte, err error) {
	reqInput = make([]byte, 0)
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if contentType == "" {
		return nil, Error_Content_Type_Required

	}
	if strings.Contains(contentType, "application/json") {
		s, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body = io.NopCloser(bytes.NewReader(s)) // 重新生成可读对象
		if !gjson.ValidBytes(s) {
			err = errors.Errorf("body content is invalid json")
			return nil, err
		}
		reqInput = s
	}
	err = r.ParseForm()
	if err != nil {
		return nil, err
	}
	if strings.Contains(contentType, "multipart/form-data") {
		err = r.ParseMultipartForm(32 << 20) // 32 MB
		if err != nil {
			return nil, err
		}
	}

	for k, values := range r.Form { // 收集表单数据
		value := ""
		if len(values) > 0 {
			value = values[0]
		}
		reqInput, err = sjson.SetBytes(reqInput, k, value)
		if err != nil {
			return nil, err
		}
	}

	for k, arr := range r.URL.Query() {
		var value any
		if useArrInQueryAndHead {
			value = arr
		} else {
			value = arr[0]
		}
		reqInput, err = sjson.SetBytes(reqInput, k, value)
		if err != nil {
			return nil, err
		}
	}

	scheme := "http"
	if strings.Contains(strings.ToLower(r.Proto), "https") {
		scheme = "https"
	}
	u := url.URL{
		Scheme: scheme,
		Path:   r.URL.Path,
		Host:   r.Host,
	}
	reqInput, err = sjson.SetBytes(reqInput, "http_url", u.String())
	if err != nil {
		return nil, err
	}
	reqInput, err = sjson.SetBytes(reqInput, "content-type", contentType)
	if err != nil {
		return nil, err
	}
	return reqInput, nil
}

func FillterAuth(w http.ResponseWriter, r *http.Request) (err error) {
	authKey := auth.GetAuthKey()
	var token string
	token = r.Header.Get(authKey)
	if token == "" {
		cooke, err := r.Cookie(authKey)
		if err == nil { // cookie 存在赋值
			token = cooke.Value
		}
	}
	if token == "" {
		token = r.PostFormValue(authKey)
	}
	if token == "" {
		token = r.FormValue(authKey)
	}

	authFunc, ok := auth.GetAuthFunc()
	if !ok {
		err = errors.New("not found authFunc,please call auth.RegisterAuthFunc before")
		return err
	}
	user, err := authFunc(token)
	if err != nil {
		return err
	}

	// 修改请求，增加auth.USER_ID_KEY 参数
	if r.URL.RawQuery != "" {
		r.URL.RawQuery = fmt.Sprintf("&%s", r.URL.RawQuery)
	}
	r.URL.RawQuery = fmt.Sprintf("%s=%s", auth.USER_ID_KEY, user.GetId()) // 增加userId
	if r.Form == nil {
		r.Form = url.Values{}
	}
	r.Form.Add(auth.USER_ID_KEY, user.GetId())
	return nil
}

func ErrorHandlerFn() (handlerErrFn stream.HandlerErrorFn) {
	return func(ctx context.Context, err error) (out []byte) {
		e := ErrorOut{
			Code:    "1",
			Message: err.Error(),
		}
		out, err1 := json.Marshal(e)
		if err1 != nil {
			panic(err1)
		}
		return out
	}
}

type ErrorOut struct {
	Code    string `json:"code"`    // 业务状态码
	Message string `json:"message"` // 业务提示
}
