package onebehaviorentity

import (
	"encoding/json"

	"github.com/suifengpiao14/errorformatter"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/xeipuuv/gojsonschema"
)

/**
  * 分多个接口，主要是方便指导使用者按顺序调用
**/

type OnebehaviorentityInterface interface {
	//Build set attribute,out linejsonschema,and _errChain, entity should be ref to change attribute value, pure function
	Build(entity OnebehaviorentityInInterface, attrSchema string, outSchema string, doFn func() (out interface{}, err error)) (stepIn OnebehaviorentityInInterface)
	//InJsonSchema get attr validate jsonschema,pure function
	InJsonSchema() (jsonschema string, err error)
	ErrorInterface
}

type OnebehaviorentityInInterface interface {
	//In set input  pure function
	In(input []byte) (stepDo OnebehaviorentityDoInterface)
	//Error get error
	ErrorInterface
}
type OnebehaviorentityDoInterface interface {
	//Do exec logic,have side effect
	Do() (stepOut OnebehaviorentityOutInterface)
	//Error get error
	ErrorInterface
}

type OnebehaviorentityOutInterface interface {
	//Out get output,pure function
	Out(out interface{}) (errInterface ErrorInterface)
}

type ErrorInterface interface {
	//Error get error
	Error() (err error)
}

func NewOnebehaviorentity() OnebehaviorentityInterface {
	return &Onebehaviorentity{}
}

type Onebehaviorentity struct {
	input      []byte
	attrSchema string
	out        interface{}
	outSchema  string
	_entity    OnebehaviorentityInInterface
	_errChain  errorformatter.ErrorChain
	_isDone    bool
	_doFn      func() (out interface{}, err error)
}

//Build 初始化实体，封装 输入输出验证格式，纯函数
func (h *Onebehaviorentity) Build(entity OnebehaviorentityInInterface, attrSchema string, outSchema string, doFn func() (out interface{}, err error)) (stepIn OnebehaviorentityInInterface) {
	h.attrSchema = attrSchema
	h.outSchema = outSchema
	h._entity = entity
	h._doFn = doFn
	h._errChain = errorformatter.NewErrorChain()
	return h
}

//In 接收参数，并且验证参数，是纯函数，和Do 分开，方便批量提前验证入参，之后异步执行Do方法
func (h *Onebehaviorentity) In(input []byte) (stepDo OnebehaviorentityDoInterface) {
	h.input = input
	h.validatInput()
	if h._errChain.Error() != nil {
		return h
	}
	if input == nil { // ignore input
		return h
	}

	// set h._entity attribute
	h._errChain.SetError(json.Unmarshal(h.input, h._entity))
	return h
}

//Do 执行业务逻辑，可能有副作用操作(数据存储),所以和Out分开
func (h *Onebehaviorentity) Do() (stepOut OnebehaviorentityOutInterface) {
	if h._errChain.Error() != nil {
		return h
	}
	out, err := h._doFn() // call h._entity Do
	h._errChain.SetError(err)
	h.out = out
	return h
}

//Out 获取返回，纯函数，和Do分开，其一从Do中提取纯函数部分，其二对有些不关心返回结果的Do省略输出转换步骤
func (h *Onebehaviorentity) Out(out interface{}) (errInterface ErrorInterface) {
	h.validateOutput()
	if h.Error() != nil {
		return h
	}
	if out == nil { //ignore output
		return h
	}
	b, err := json.Marshal(h.out)
	if err != nil {
		h._errChain.SetError(err)
		return h
	}
	h._errChain.SetError(json.Unmarshal(b, out))
	return h
}

func (h *Onebehaviorentity) Error() (err error) {
	return h._errChain.Error()
}

func (h *Onebehaviorentity) InJsonSchema() (schema string, err error) {
	lineSchema, err := jsonschemaline.ParseJsonschemaline(h.attrSchema)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	b, err := lineSchema.JsonSchema()
	schema = string(b)
	if err != nil {
		h._errChain.SetError(err)
		return "", err
	}
	return schema, nil
}

func (h *Onebehaviorentity) validatInput() {
	if h.Error() != nil {
		return
	}
	if h.attrSchema == "" {
		return
	}
	lineSchema, err := jsonschemaline.ParseJsonschemaline(h.attrSchema)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	jsb, err := lineSchema.JsonSchema()
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	jsonschemaStr := string(jsb)
	schemaLoader := gojsonschema.NewStringLoader(jsonschemaStr)
	err = util.Validate(string(h.input), schemaLoader)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
}

func (h *Onebehaviorentity) validateOutput() {
	if h.Error() != nil {
		return
	}
	if h.outSchema == "" {
		return
	}
	lineSchema, err := jsonschemaline.ParseJsonschemaline(h.outSchema)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	jsb, err := lineSchema.JsonSchema()
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	jsonschemaStr := string(jsb)
	schemaLoader := gojsonschema.NewStringLoader(jsonschemaStr)
	out, err := json.Marshal(h.out)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
	err = util.Validate(string(out), schemaLoader)
	if err != nil {
		h._errChain.SetError(err)
		return
	}
}
