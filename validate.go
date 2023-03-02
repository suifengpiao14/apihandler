package controllerhandler

import (
	"github.com/pkg/errors"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/xeipuuv/gojsonschema"
)

type ValidateI interface {
	GetLineSchema() (lineschema string)
}

// ValidateI 接口函数式实现
type ValidateIFn func() (lineschema string)

func (fn ValidateIFn) GetLineSchema() (lineschema string) {
	return fn()
}

type Validate struct {
	ValidateI
	_schemaLoader gojsonschema.JSONLoader
}

func NewValidate(validateI ValidateI) (validate *Validate, err error) {
	validate = &Validate{
		ValidateI: validateI,
	}
	if err = validate.init(); err != nil {
		return nil, err
	}

	return validate, nil
}

func (v *Validate) init() (err error) {
	lineSchemaStr := v.GetLineSchema()
	if lineSchemaStr == "" {
		return
	}
	inputlineSchema, err := jsonschemaline.ParseJsonschemaline(lineSchemaStr)
	if err != nil {
		return err
	}
	if inputlineSchema != nil {
		jsb, err := inputlineSchema.JsonSchema()
		if err != nil {
			return err
		}
		jsonschemaStr := string(jsb)
		v._schemaLoader = gojsonschema.NewStringLoader(jsonschemaStr)
	}
	return
}

func (v Validate) Validate(jsonStr string) (err error) {
	if v._schemaLoader == nil {
		return nil
	}
	err = util.Validate(jsonStr, v._schemaLoader)
	if err != nil {
		return err
	}
	return nil
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
