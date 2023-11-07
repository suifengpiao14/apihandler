package example

import (
	"context"
)

func (i *AdListInput) Do(ctx context.Context) (err error) {
	input := i
	output := &input.Output
	output.Items = make([]AdListOutputItem, 0)
	output.Pagination.Index = input.Index
	output.Pagination.Size = input.Size
	output.Pagination.Total = 100
	item := AdListOutputItem{
		ID: 1,
	}
	output.Items = append(output.Items, item)
	return nil
}
