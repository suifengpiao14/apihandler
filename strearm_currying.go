package apihandler

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/gojsonschemavalidator"
	"github.com/suifengpiao14/jsonschemaline"
	"github.com/suifengpiao14/stream"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
)

func MakeMergeDefaultHandler(defaultJson []byte) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		if defaultJson == nil {
			return input, nil
		}
		inputStr, err := jsonschemaline.MergeDefault(string(input), string(defaultJson))
		if err != nil {
			err = errors.WithMessage(err, "merge default value error")
			return nil, err
		}

		return []byte(inputStr), nil
	}
}

func MakeValidateHandler(validateLoader gojsonschema.JSONLoader) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		if validateLoader == nil {
			return input, nil
		}
		inputStr := string(input)
		err = gojsonschemavalidator.Validate(inputStr, validateLoader)
		if err != nil {
			return nil, err
		}
		return input, nil
	}
}

func MakeFormatHandler(formatGjsonPath string) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		formattedInput := gjson.Get(string(input), formatGjsonPath).String()
		out = []byte(formattedInput)
		return out, nil
	}
}

func MakeUnmarshalHandler(dst interface{}) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		err = json.Unmarshal([]byte(input), dst)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func MakeUnPackHandler() (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		return out, nil
	}
}

func MakePackHandler() (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		outStr, err := jsonschemaline.MergeDefault(string(input), `{"code":"0","message":"ok"}`)
		if err != nil {
			return nil, err
		}
		out = []byte(outStr)
		return out, nil
	}
}
