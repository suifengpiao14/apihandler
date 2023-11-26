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

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/suifengpiao14/apihandler/auth"
	"github.com/suifengpiao14/stream"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var API_NOT_FOUND = errors.Errorf("not found API")
var (
	ERROR_NOT_IMPLEMENTED = errors.New("not implemented")
)

type ApiInterface interface {
	GetRoute() (method string, path string)
	Init()
	GetDescription() (title string, description string)
	GetName() (domain string, name string)
	SetContext(ctx context.Context)
	GetContext() (ctx context.Context)
	Run(input []byte) (out []byte, err error)
	Do(ctx context.Context) (err error)
	GetOutRef() (outRef OutI)
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

func (e *DefaultImplementFuncs) Init() {
}

func (e *DefaultImplementFuncs) SetContext(ctx context.Context) {
	e.ctx = ctx

}
func (e *DefaultImplementFuncs) GetContext() (ctx context.Context) {
	return e.ctx
}

type OutI interface {
	Bytes() (out []byte)
}

type OutputString string

func (output *OutputString) Bytes() (out []byte) {
	out = []byte(*output)
	return out
}

type _OutputJson struct {
	v any
}

func (output _OutputJson) Bytes() (out []byte) {
	b, err := json.Marshal(output.v)
	if err != nil {
		s := fmt.Sprintf("{message:%s}", err.Error())
		out = []byte(s)
		return
	}
	return b
}

func OutputJson(v any) OutI {
	return _OutputJson{
		v: v,
	}
}

func JsonMarshalOutput(o interface{}) (out []byte) {
	b, err := json.Marshal(o)
	if err != nil {
		s := fmt.Sprintf("{message:%s}", err.Error())
		out = []byte(s)
		return
	}
	out = b
	return out
}

var apiMap sync.Map

// //LineschemaPacketStream lineschema 包 流处理函数
// func LineschemaPacketStream(api ApiInterface, lineschemaApi lineschemapacket.LineschemaPacketI) (s *stream.Stream, err error) {

// 	lineschemaPacketHandlers, err := lineschemapacket.ServerPackHandlers(lineschemaApi)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s = stream.NewStream(api.ErrorHandle, lineschemaPacketHandlers...)
// 	s.AddPack(stream.Bytes2Stuct2BytesJsonPacket(api, api.GetOutRef()), wrapDo(api))
// 	return s, err
// }

// ApiPackHandlers api 处理流函数
func ApiPackHandlers(api ApiInterface) (packHandlers stream.PackHandlers) {
	packHandlers = make(stream.PackHandlers, 0)
	packHandlers.Add(
		stream.Bytes2Stuct2BytesJsonPacket(api, api.GetOutRef()),
		wrapDo(api),
	)
	return packHandlers
}

// wrapDo 把api.Do函数柯里化
func wrapDo(api ApiInterface) stream.PackHandler {
	return stream.NewPackHandler(func(ctx context.Context, _ []byte) (_ []byte, err error) {
		err = api.Do(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}, nil)

}

type ApiKey struct {
	Method string
	Path   string
}

func (rk ApiKey) String() (s string) {
	s = fmt.Sprintf("%s####%s", rk.Method, rk.Path)
	return s
}

func NewApiKey(method string, path string) (k ApiKey) {
	return ApiKey{
		Method: method,
		Path:   path,
	}
}

// RegisterApi 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func RegisterApi(apiInterface ApiInterface) (err error) {
	method, path := apiInterface.GetRoute()
	key := NewApiKey(method, path)
	v, ok := apiMap.Load(key)
	if ok {
		err = errors.Errorf("api key already registered,key:%s,value:%T", key, v)
		return err
	}
	apiMap.Store(key, apiInterface)
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
		method, path := route.Method, route.Path
		routeFn(method, path)
	}
}

func getAllAPI() (apis []ApiInterface, err error) {
	routes := GetAllRoute()
	apis = make([]ApiInterface, 0)
	for _, route := range routes {
		method, path := route.Method, route.Path
		api, err := GetApi(context.Background(), method, path)
		if err != nil {
			return nil, err
		}
		apis = append(apis, api)
	}
	return apis, nil
}

// GetAllRoute 获取已注册的所有api route
func GetAllRoute() (apiKeys []ApiKey) {
	apiKeys = make([]ApiKey, 0)
	apiMap.Range(func(key, value any) bool {
		apiKey, ok := key.(ApiKey)
		if ok {
			apiKeys = append(apiKeys, apiKey)
		}
		return true
	})
	return apiKeys
}

func GetApi(ctx context.Context, method string, path string) (api ApiInterface, err error) {
	key := NewApiKey(method, path)
	apiAny, ok := apiMap.Load(key)
	if !ok {
		return nil, errors.WithMessagef(API_NOT_FOUND, "method:%s,path:%s", method, path)
	}
	exitsApi := apiAny.(ApiInterface)
	rt := reflect.TypeOf(exitsApi).Elem()
	rv := reflect.New(rt)
	api = rv.Interface().(ApiInterface)
	api.Init()
	api.SetContext(ctx)
	return api, nil
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
