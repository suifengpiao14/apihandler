package apihandler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/xeipuuv/gojsonschema"
)

type ApiInterface interface {
	GetDoFn() func(ctx context.Context) (out OutputI, err error)
	GetInputSchema() (lineschema string)
	GetOutputSchema() (lineschema string)
}

type OutputI interface {
	String() (out string, err error)
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

// NewApi 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func NewApi(apiInterface ApiInterface) (api *Api, err error) {
	rt := reflect.TypeOf(apiInterface)
	kind := rt.Kind()
	if kind != reflect.Ptr {
		err = errors.Errorf("want:Ptr,got:%s", kind)
		return nil, err
	}
	rtE := rt.Elem()
	key := fmt.Sprintf("%s.%s", rtE.PkgPath(), rtE.Name())
	if apiI, ok := apiMap.Load(key); ok {
		exitsApi := apiI.(*Api)
		api = &Api{ApiInterface: apiInterface, validateInputLoader: exitsApi.validateInputLoader, validateOutputLoader: exitsApi.validateOutputLoader}
		return api, nil
	}

	// 以下初始化可以复用,线程安全
	api = &Api{
		ApiInterface: apiInterface,
	}
	inputSchema := apiInterface.GetInputSchema()
	if inputSchema != "" {
		api.validateInputLoader, err = NewJsonschemaLoader(inputSchema)
		if err != nil {
			return nil, err
		}
	}
	outputSchema := apiInterface.GetOutputSchema()
	if outputSchema != "" {
		api.validateOutputLoader, err = NewJsonschemaLoader(outputSchema)
		if err != nil {
			return nil, err
		}
	}
	apiMap.Store(key, api)
	return api, nil
}

func (a Api) inputValidate(input string) (err error) {
	inputStr := string(input)
	err = util.Validate(inputStr, a.validateInputLoader)
	if err != nil {
		return err
	}
	return nil
}
func (a Api) outputValidate(output string) (err error) {
	outputStr := string(output)
	err = util.Validate(outputStr, a.validateOutputLoader)
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