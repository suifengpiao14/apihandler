package controllerhandler

import (
	"context"
	"encoding/json"
)

type HandlerInterface interface {
	GetLineSchemaInput() (lineschema string)
	GetLineSchemaOutput() (lineschema string)
	Do(ctx context.Context) (out OutputI, err error)
}

type OutputI interface {
	String() (out string, err error)
}

func JsonMarshal(o interface{}) (out string, err error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	out = string(b)
	return out, nil
}

type Handler struct {
	HandlerInterface
	validateInput  *Validate
	validateOutput *Validate
}

//NewHandler 创建处理器，内部逻辑在接收请求前已经确定，后续不变，所以有错误直接panic ，能正常启动后，这部分不会出现错误
func NewHandler(handlerInterface HandlerInterface) (handler *Handler) {
	var inputValidateI ValidateIFn = func() (lineschema string) {
		return handlerInterface.GetLineSchemaInput()
	}
	validateInput, err := NewValidate(inputValidateI)
	if err != nil {
		panic(err)
	}

	var outputValidateI ValidateIFn = func() (lineschema string) {
		return handlerInterface.GetLineSchemaOutput()
	}
	validateOutput, err := NewValidate(outputValidateI)
	if err != nil {
		panic(err)
	}

	handler = &Handler{HandlerInterface: handlerInterface, validateInput: validateInput, validateOutput: validateOutput}
	if err != nil {
		panic(err)
	}

	return handler
}

func (a Handler) inputValidate(input string) (err error) {
	inputStr := string(input)
	err = a.validateInput.Validate(inputStr)
	if err != nil {
		return err
	}
	return nil
}
func (a Handler) outputValidate(output string) (err error) {
	outputStr := string(output)
	err = a.validateOutput.Validate(outputStr)
	if err != nil {
		return err
	}
	return nil
}

func (a Handler) Run(ctx context.Context, input string) (out string, err error) {
	err = a.inputValidate(input)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(input), a.HandlerInterface)
	if err != nil {
		return "", err
	}
	outI, err := a.HandlerInterface.Do(ctx)
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
