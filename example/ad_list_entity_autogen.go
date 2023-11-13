package example

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/suifengpiao14/apihandler"
	"github.com/suifengpiao14/lineschemapacket"
	"github.com/suifengpiao14/stream"
)

func init() {
	err := apihandler.RegisterApi(&AdListInput{})
	if err != nil {
		panic(err)
	}
	err = lineschemapacket.RegisterLineschemaPacket(&AdListInput{})
	if err != nil {
		panic(err)
	}
}

type AdListInput struct {
	Title        string       `json:"title"`
	AdvertiserID int          `json:"advertiserId"`
	BeginAt      string       `json:"beginAt"`
	EndAt        string       `json:"endAt"`
	Index        int          `json:"index"`
	Size         int          `json:"size"`
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
func (i *AdListInput) UnpackSchema() (lineschema string) {
	lineschema = `
	version=http://json-schema.org/draft-07/schema#,id=in,direction=in
	fullname=title,required,description=广告标题,comment=广告标题,example=新年豪礼
	fullname=advertiserId,format=int,required,description=广告主,comment=广告主,example=123
	fullname=beginAt,required,description=可以投放开始时间,comment=可以投放开始时间,example=2023-01-12 00:00:00
	fullname=endAt,required,description=投放结束时间,comment=投放结束时间,example=2023-01-30 00:00:00
	fullname=index,required,format=int,description=页索引,0开始,default=0,comment=页索引,0开始
	fullname=size,required,format=int,description=每页数量,default=10,comment=每页数量
	fullname=appid,required,description=访问服务的备案id,comment=访问服务的备案id
	fullname=signature,required,description=签名,外网访问需开启签名,comment=签名,外网访问需开启签名
	`
	return
}

func (i *AdListInput) PackSchema() (lineschema string) {
	lineschema = `
	version=http://json-schema.org/draft-07/schema#,id=out,direction=out
	fullname=code,description=业务状态码,comment=业务状态码,default=0,example=0
	fullname=message,description=业务提示,comment=业务提示,default=ok,example=ok
	fullname=items,type=array,description=数组,comment=数组,example=-
	fullname=items[].id,format=int,description=主键,comment=主键,example=0
	fullname=items[].title,description=广告标题,comment=广告标题,example=新年豪礼
	fullname=items[].advertiserId,description=广告主,comment=广告主,example=123
	fullname=items[].summary,description=广告素材-文字描述,comment=广告素材-文字描述,example=下单有豪礼
	fullname=items[].image,description=广告素材-图片地址,comment=广告素材-图片地址
	fullname=items[].link,description=连接地址,comment=连接地址
	fullname=items[].type,description=广告素材(类型),text-文字,image-图片,vido-视频,comment=广告素材(类型),text-文字,image-图片,vido-视频,example=image
	fullname=items[].beginAt,description=投放开始时间,comment=投放开始时间,example=2023-01-12 00:00:00
	fullname=items[].endAt,description=投放结束时间,comment=投放结束时间,example=2023-01-30 00:00:00
	fullname=items[].remark,description=备注,comment=备注,example=营养早餐广告
	fullname=items[].valueObj,description=json扩展,广告的值属性对象,comment=json扩展,广告的值属性对象,example={"tag":"index"}
	fullname=pagination,type=object,description=对象,comment=对象
	fullname=pagination.index,format=int,description=页索引,0开始,comment=页索引,0开始,example=0
	fullname=pagination.size,format=int,description=每页数量,comment=每页数量,example=10
	fullname=pagination.total,format=int,description=总数,comment=总数,example=60
	`
	return
}

func (i *AdListInput) GetOutRef() (out apihandler.OutI) {
	return &i.Output
}

func (i *AdListInput) Run(input []byte) (out []byte, err error) {
	lineschemaPacketHandlers, err := lineschemapacket.ServerPackHandlers(i)
	if err != nil {
		return nil, err
	}
	s := stream.NewStream(i.ErrorHandle, lineschemaPacketHandlers...)
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
