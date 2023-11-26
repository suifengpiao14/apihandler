package example

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/suifengpiao14/apihandler"
	"github.com/suifengpiao14/stream"
)

func init() {
	err := apihandler.RegisterApi(&AdListInput{})
	if err != nil {
		panic(err)
	}

}

type AdListInput struct {
	Title        string       `json:"title"`
	AdvertiserID int          `json:"advertiserId,string"`
	BeginAt      string       `json:"beginAt"`
	EndAt        string       `json:"endAt"`
	Index        int          `json:"index,string"`
	Size         int          `json:"size,string"`
	Output       AdListOutput `json:"-"`
	apihandler.DefaultImplementFuncs
}

type AdListOutput struct {
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	Items      []AdListOutputItem `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

type AdListOutputItem struct {
	ID           int    `json:"id"`
	Title        string `json:"title,omitempty"`
	AdvertiserID string `json:"advertiserId"`
	Summary      string `json:"summary"`
	Image        string `json:"image"`
	Link         string `json:"link"`
	Type         string `json:"type"`
	BeginAt      string `json:"beginAt"`
	EndAt        string `json:"endAt"`
	Remark       string `json:"remark"`
	ValueObj     string `json:"valueObj"`
}

type Pagination struct {
	Index int `json:"index,string"`
	Size  int `json:"size,string"`
	Total int `json:"total,string"`
}

func (o AdListOutput) Bytes() (out []byte) {
	return apihandler.JsonMarshalOutput(o)
}

func (i *AdListInput) Init() {
}

func (e *AdListInput) GetName() (domain string, name string) {
	return "example", "adList"
}
func (e *AdListInput) GetDescription() (title string, description string) {
	return "广告列表", "广告列表"
}

func (i *AdListInput) GetRoute() (method string, path string) {
	path = "/api/v1/adList"
	return http.MethodPost, path
}

func (i *AdListInput) GetOutRef() (out apihandler.OutI) {
	return &i.Output
}

func (i *AdListInput) Run(input []byte) (out []byte, err error) {
	s := stream.NewStream(i.ErrorHandle)
	s.AddPack(apihandler.ApiPackHandlers(i)...)
	ctx := i.GetContext()
	out, err = s.Run(ctx, input)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (i *AdListInput) ErrorHandle(ctx context.Context, err error) (out []byte) {
	e := ErrorOut{
		Code:    "1",
		Message: err.Error(),
	}
	out, err1 := json.Marshal(e)
	if err1 != nil {
		panic(err1)
	}
	return out
}

type ErrorOut struct {
	Code    string `json:"code"`    // 业务状态码
	Message string `json:"message"` // 业务提示
}
