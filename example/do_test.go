package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/logchan/v2"
)

func TestRun(t *testing.T) {
	input := `{"title":"新年豪礼","advertiserId":"123","beginAt":"2023-01-12 00:00:00","endAt":"2023-01-30 00:00:00","index":"0","size":"10","content-type":"application/json","appid":"","signature":""}`
	ctx := context.Background()
	out, err := Run(ctx, input)
	require.NoError(t, err)
	fmt.Println(out)
	logchan.CloseLogChan()
}
