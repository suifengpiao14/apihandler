package apihandler

import (
	"context"
	"encoding/json"

	"github.com/suifengpiao14/stream"
)

func MakeUnmarshalFn(dst interface{}) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		if input == nil {
			return nil, nil
		}
		err = json.Unmarshal([]byte(input), dst)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func MakeUnPackFn() (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		return input, nil
	}
}
func MakePackFn() (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		return input, nil
	}
}


func MakeDoFn(api ApiInterface) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		if input == nil {
			return nil, nil
		}
		err = json.Unmarshal([]byte(input), api)
		if err != nil {
			return nil, err
		}
		doFn := api.GetDoFn()
		outI, err := doFn(ctx)
		if err != nil {
			return nil, err
		}
		outStr := outI.String()
		out = []byte(outStr)
		return out, nil
	}
}
