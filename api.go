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
	"github.com/suifengpiao14/gojsonschemavalidator"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
)

var API_NOT_FOUND = errors.Errorf("not found")

type ApiInterface interface {
	GetDoFn() func(ctx context.Context) (out OutputI, err error)
	GetInputSchema() (lineschema string)
	GetOutputSchema() (lineschema string)
	GetRoute() (method string, path string)
}

type OutputI interface {
	String() (out string, err error)
}

type OutputString string

func (output *OutputString) String() (out string, err error) {
	out = string(*output)
	return out, nil
}

func GetApiInterfaceID(apiInterface ApiInterface) (id string) {
	rt := reflect.TypeOf(apiInterface)
	kind := rt.Kind()
	if kind != reflect.Ptr {
		err := errors.Errorf("want:Ptr,got:%s", kind)
		panic(err)
	}
	rtE := rt.Elem()
	id = fmt.Sprintf("%s.%s", rtE.PkgPath(), rtE.Name())
	return id
}

func GetRouteKey(method string, path string) (key string) {
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

type Api struct {
	ApiInterface
	validateInputLoader  gojsonschema.JSONLoader
	validateOutputLoader gojsonschema.JSONLoader
}

var apiMap sync.Map

// RegisterApi 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func RegisterApi(apiInterface ApiInterface) (err error) {
	method, path := apiInterface.GetRoute()
	key := GetRouteKey(method, path)
	// 以下初始化可以复用,线程安全
	api := &Api{
		ApiInterface: apiInterface,
	}
	inputSchema := apiInterface.GetInputSchema()
	if inputSchema != "" {
		api.validateInputLoader, err = NewJsonschemaLoader(inputSchema)
		if err != nil {
			return err
		}
	}
	outputSchema := apiInterface.GetOutputSchema()
	if outputSchema != "" {
		api.validateOutputLoader, err = NewJsonschemaLoader(outputSchema)
		if err != nil {
			return err
		}
	}
	apiMap.Store(key, api)
	return nil
}

func Run(ctx context.Context, r *http.Request) (out string, err error) {
	method, path := r.Method, r.URL.Path
	api, err := GetApi(method, path)
	if err != nil {
		return "", err
	}
	input, err := FormatInput(r, false)
	if err != nil {
		return "", err
	}
	out, err = api.Run(ctx, string(input))
	if err != nil {
		return "", err
	}
	return out, nil

}

func GetApi(method string, path string) (api Api, err error) {
	key := GetRouteKey(method, path)
	apiAny, ok := apiMap.Load(key)
	if !ok {
		return api, errors.WithMessagef(API_NOT_FOUND, "method:%s,path:%s", method, path)
	}
	exitsApi := apiAny.(*Api)
	api = Api{ApiInterface: exitsApi.ApiInterface, validateInputLoader: exitsApi.validateInputLoader, validateOutputLoader: exitsApi.validateOutputLoader}
	return api, nil
}

func (a Api) inputValidate(input string) (err error) {
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
func (a Api) outputValidate(output string) (err error) {
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

func (a Api) convertInput(input string) (err error) {
	err = json.Unmarshal([]byte(input), a.ApiInterface)
	if err != nil {
		return err
	}
	return nil
}

func (a Api) Run(ctx context.Context, input string) (out string, err error) {

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
	err = a.convertInput(input)
	if err != nil {
		return "", err
	}
	doFn := a.ApiInterface.GetDoFn()
	outI, err := doFn(ctx)
	if err != nil {
		return "", err
	}
	out, err = outI.String()
	if err != nil {
		return "", err
	}
	err = a.outputValidate(out)
	if err != nil {
		return "", err
	}
	return out, nil
}

func NewJsonschemaLoader(lineSchemaStr string) (jsonschemaLoader gojsonschema.JSONLoader, err error) {
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

//FormatInput 统一获取 query,header,body 参数
func FormatInput(r *http.Request, useArrInQueryAndHead bool) (reqInput []byte, err error) {
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
