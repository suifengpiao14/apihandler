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
	Build(entity OnebehaviorentityInterface, attrSchema string, outSchema string)
	//In set input
	In(input []byte)
	//Out get output
	Out(out interface{}) (err error)
	//Do Implementation business logic
	Do() (out interface{}, err error)
	//InJsonSchema get attr validate jsonschema
	InJsonSchema() (jsonschema string, err error)
	//Error get error
	Error() (err error)
}

type Onebehaviorentity struct {
	input      []byte
	attrSchema string
	out        interface{}
	outSchema  string
	_entity    OnebehaviorentityInterface
	_errChain  errorformatter.ErrorChain
	_isDone    bool
}

func (h *Onebehaviorentity) Build(entity OnebehaviorentityInterface, attrSchema string, outSchema string) {
	h.attrSchema = attrSchema
	h.outSchema = outSchema
	h._entity = entity
	h._errChain = errorformatter.NewErrorChain()
}

func (h *Onebehaviorentity) In(input []byte) {
	h.input = input
	h.validatInput()
	if h.Error() == nil {
		// set h._entity attribute
		h._errChain.SetError(json.Unmarshal(h.input, h._entity))
	}
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

func (h *Onebehaviorentity) Out(out interface{}) (err error) {
	if !h._isDone {
		h._isDone = true
		tmpOut, err := h._entity.Do() // call h._entity Do
		if err != nil {
			h._errChain.SetError(err)
		}
		h.out = tmpOut

	}
	h.validateOutput()
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

func (h *Onebehaviorentity) Do() (out interface{}, err error) {

	//h._entity  Implementation business logic
	return
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
