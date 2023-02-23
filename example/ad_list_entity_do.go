package example

import (
	"context"

	"github.com/suifengpiao14/controllerhandler"
)

func AdListDoFn(ctx context.Context, i AdListInput) (controllerhandler.OutputI, error) {
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
