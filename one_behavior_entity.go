package Onebehaviorentity

import (
	"encoding/json"

	"github.com/suifengpiao14/errorformatter"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/xeipuuv/gojsonschema"
)

type OnebehaviorentityInterface interface {
	//SetSchema set attribute,out linejsonschema
	SetSchema(attrSchema string, outSchema string)
	//In set input
	In(input []byte)
	//Out get output
	Out(out interface{}) (err error)
	//ValidatInput validate input
	ValidatInput()
	//Do Implementation business logic
	Do()
	//ValidateOutput validate output schema
	ValidateOutput()
	//JsonSchema get attr validate jsonschema
	JsonSchema() (jsonschema string, err error)
	//Error get error
	Error() (err error)
}

type Onebehaviorentity struct {
	input      []byte
	attrSchema string
	out        interface{}
	outSchema  string
	_errChain  errorformatter.ErrorChain
	_isDone    bool
}

func (h *Onebehaviorentity) SetSchema(attrSchema string, outSchema string) {
	h.attrSchema = attrSchema
	h.outSchema = outSchema
}

func (h *Onebehaviorentity) In(input []byte) {
	h.input = input
	h.ValidatInput()
	if h.Error() == nil {
		h._errChain.SetError(json.Unmarshal(h.input, h)) // 当前实现,没有对外属性,被嵌入后,主体可能提供公共属性
	}
}

func (h *Onebehaviorentity) JsonSchema() (schema string, err error) {
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

func (h *Onebehaviorentity) Error() (err error) {
	return h._errChain.Error()
}

func (h *Onebehaviorentity) Out(out interface{}) (err error) {
	if !h._isDone {
		h.Do()
	}
	h.ValidateOutput()
	err = h.Error()
	if err != nil {
		return err
	}

	b, err := json.Marshal(h.out)
	if err != nil {
		h._errChain.SetError(err)
		return err
	}
	h._errChain.SetError(json.Unmarshal(b, out))
	return nil
}

func (h *Onebehaviorentity) Do() {
	if h.Error() != nil {
		return
	}
	h._isDone = true
	//todo Implementation business logic
}

func (h *Onebehaviorentity) ValidatInput() {
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

func (h *Onebehaviorentity) ValidateOutput() {
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
