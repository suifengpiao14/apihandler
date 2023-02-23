package example

import (
	"context"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/controllerhandler"
)

func AdListDoFn(ctx context.Context, handler controllerhandler.Handler) (controllerhandler.OutputI, error) {
	i, ok := handler.HandlerInterface.(*AdListInput)
	if !ok {
		err := errors.Errorf("handler must be AdListInput")
		return nil, err
	}
	output := i.Output
	output.Code = "200"
	output.Message = "ok"
	output.Items = make([]AdListOutputItem, 0)
	output.Pagination.Index = i.Index
	output.Pagination.Size = i.Size
	output.Pagination.Total = 100
	item := AdListOutputItem{
		ID: "1",
	}
	output.Items = append(output.Items, item)
	return output, nil
}
