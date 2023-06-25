package apihandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/gojsonschemavalidator"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
)

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
}

type EmptyApi struct{}

func (e *EmptyApi) GetDoFn() (doFn func(ctx context.Context) (out OutputI, err error)) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetDoFn")
	panic(err)
}
func (e *EmptyApi) GetInputSchema() (lineschema string) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetInputSchema")
	panic(err)
}
func (e *EmptyApi) GetOutputSchema() (lineschema string) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetOutputSchema")
	panic(err)
}
func (e *EmptyApi) GetRoute() (method string, path string) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetRoute")
	panic(err)
}
func (e *EmptyApi) Init() {
}

type OutputI interface {
	String() (out string, err error)
}

type OutputString string

func (output *OutputString) String() (out string, err error) {
	out = string(*output)
	return out, nil
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

type _Api struct {
	ApiInterface
	inputFormatGjsonPath  string
	outputFormatGjsonPath string
	validateInputLoader   gojsonschema.JSONLoader
	validateOutputLoader  gojsonschema.JSONLoader
}

var apiMap sync.Map

const (
	apiMap_route_add_key = "___all_route_add___"
	apiMap_route_del_key = "___all_route_del___"
)

// RegisterApi 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func RegisterApi(apiInterface ApiInterface) (err error) {
	method, path := apiInterface.GetRoute()
	key := getRouteKey(method, path)
	// 以下初始化可以复用,线程安全
	api := &_Api{
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

func RegisterRouteFn(routeFn func(method string, path string)) {
	routes := GetAllRoute()
	for _, route := range routes {
		method, path := route[0], route[1]
		routeFn(method, path)
	}
}

// GetAllRoute 获取已注册的所有api route
func GetAllRoute() (routes [][2]string) {
	routes = make([][2]string, 0)
	delRouteMap := getAllDelRoute()
	if routesI, ok := apiMap.Load(apiMap_route_add_key); ok {
		if tmp, ok := routesI.(map[string][2]string); ok {
			for key, route := range tmp {
				if _, ok := delRouteMap[key]; ok {
					continue
				}
				routes = append(routes, route)
			}
		}
	}

	return routes
}

// RemoveRoute 记录删除的api 路由(部分路由可能已经注册，GetAllRoute 会排除)
func RemoveRoute(method string, path string) {
	key := getRouteKey(method, path)
	delRoutes := make(map[string][2]string)
	if delRoutesI, ok := apiMap.Load(apiMap_route_del_key); ok {
		if old, ok := delRoutesI.(map[string][2]string); ok {
			delRoutes = old
		}
	}
	delRoutes[key] = [2]string{method, path}
	apiMap.Store(apiMap_route_del_key, delRoutes)
}

func getAllDelRoute() (routes map[string][2]string) {
	if routesI, ok := apiMap.Load(apiMap_route_del_key); ok {
		if routes, ok = routesI.(map[string][2]string); ok {
			return routes
		}
	}
	return routes
}

func Run(ctx context.Context, method string, path string, input string) (out string, err error) {
	api, err := GetApi(method, path)
	if err != nil {
		return "", err
	}
	out, err = api.Run(ctx, string(input))
	if err != nil {
		return "", err
	}
	return out, nil

}

func GetApi(method string, path string) (api _Api, err error) {
	key := getRouteKey(method, path)
	apiAny, ok := apiMap.Load(key)
	if !ok {
		return api, errors.WithMessagef(API_NOT_FOUND, "method:%s,path:%s", method, path)
	}
	exitsApi := apiAny.(*_Api)
	rt := reflect.TypeOf(exitsApi.ApiInterface).Elem()
	rv := reflect.New(rt)
	apiInterface := rv.Interface().(ApiInterface)
	apiInterface.Init()
	api = _Api{
		ApiInterface:          apiInterface,
		validateInputLoader:   exitsApi.validateInputLoader,
		validateOutputLoader:  exitsApi.validateOutputLoader,
		inputFormatGjsonPath:  exitsApi.inputFormatGjsonPath,
		outputFormatGjsonPath: exitsApi.outputFormatGjsonPath,
	}
	return api, nil
}

func (a _Api) inputValidate(input string) (err error) {
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
func (a _Api) outputValidate(output string) (err error) {
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

func (a _Api) modifyTypeByFormat(input string, formatGjsonPath string) (formattedInput string, err error) {
	formattedInput = input
	if formatGjsonPath == "" {
		return formattedInput, nil
	}
	formattedInput = gjson.Get(input, formatGjsonPath).String()
	return formattedInput, nil
}

func (a _Api) convertInput(input string) (err error) {
	err = json.Unmarshal([]byte(input), a.ApiInterface)
	if err != nil {
		return err
	}
	return nil
}

func (a _Api) Run(ctx context.Context, input string) (out string, err error) {

	if a.ApiInterface == nil {
		err = errors.Errorf("handlerInterface required %v", a)
		return "", err
	}
	if a.ApiInterface.GetDoFn() == nil { //此处只先判断,不取值,等后续将input值填充后再获取
		err = errors.Errorf("doFn required %v", a.ApiInterface)
		return "", err
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
	err = a.convertInput(formattedInput)
	if err != nil {
		return "", err
	}
	doFn := a.ApiInterface.GetDoFn()
	outI, err := doFn(ctx)
	if err != nil {
		return "", err
	}
	if funcs.IsNil(outI) {
		err = errors.New("response not be nil ")
		err = errors.WithMessage(err, "github.com/suifengpiao14/apihandler._Api.Run")
		return "", err
	}
	originalOut, err := outI.String()
	if err != nil {
		return "", err
	}
	out, err = a.modifyTypeByFormat(originalOut, a.outputFormatGjsonPath)
	if err != nil {
		return "", err
	}
	err = a.outputValidate(out)
	if err != nil {
		return "", err
	}
	return out, nil
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

// RequestInputToJson 统一获取 query,header,body 参数
func RequestInputToJson(r *http.Request, useArrInQueryAndHead bool) (reqInput []byte, err error) {
	reqInput = make([]byte, 0)
	s, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return
	}
	if len(s) > 0 {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "application/json") {
			if !gjson.ValidBytes(s) {
				err = errors.Errorf("body content is invalid json")
				return nil, err
			}
			reqInput = s
		} else {
			reqInput, err = sjson.SetBytes(reqInput, "body", s)
			if err != nil {
				return nil, err
			}
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

	headers := r.Header
	for k, v := range headers {
		key := fmt.Sprintf("http_%s", strings.ReplaceAll(strings.ToLower(k), "-", "_"))
		var value any
		if useArrInQueryAndHead {
			value = v
		} else {
			value = v[0]
		}
		reqInput, err = sjson.SetBytes(reqInput, key, value)
		if err != nil {
			return nil, err
		}
	}
	return reqInput, nil
}
