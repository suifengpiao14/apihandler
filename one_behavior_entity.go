package onebehaviorentity

import (
	"encoding/json"

	"github.com/suifengpiao14/errorformatter"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/xeipuuv/gojsonschema"
)

type OnebehaviorentityInterface interface {
	//Build set attribute,out linejsonschema,and _errChain, entity should be ref to change attribute value,
	Build(entity OnebehaviorentityInterface, attrSchema string, outSchema string, doFn func() (out interface{}, err error))
	//In set input
	In(input []byte) (behavior OnebehaviorentityInterface)
	//Out get output
	Out(out interface{}) (behavior OnebehaviorentityInterface)
	//InJsonSchema get attr validate jsonschema
	InJsonSchema() (jsonschema string, err error)
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
	_entity    OnebehaviorentityInterface
	_errChain  errorformatter.ErrorChain
	_isDone    bool
	_doFn      func() (out interface{}, err error)
}

func (h *Onebehaviorentity) Build(entity OnebehaviorentityInterface, attrSchema string, outSchema string, doFn func() (out interface{}, err error)) {
	h.attrSchema = attrSchema
	h.outSchema = outSchema
	h._entity = entity
	h._errChain = errorformatter.NewErrorChain()
}

func (h *Onebehaviorentity) In(input []byte) (out OnebehaviorentityInterface) {
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

func (h *Onebehaviorentity) Error() (err error) {
	return h._errChain.Error()
}

func (h *Onebehaviorentity) Out(out interface{}) (behavior OnebehaviorentityInterface) {
	if !h._isDone {
		h._isDone = true
		if h._doFn == nil { // return if h._doFn is nil
			return h
		}
		tmpOut, err := h._doFn() // call h._entity Do
		h._errChain.SetError(err)
		h.out = tmpOut

	}
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
