package example

import (
	"context"
	"net/http"

	"github.com/suifengpiao14/apihandler"
)

type AdListInput struct {
	Title       string       `json:"title"`
	AdvertiseID int          `json:"advertiseId,string"`
	BeginAt     string       `json:"beginAt"`
	EndAt       string       `json:"endAt"`
	Index       int          `json:"index,string"`
	Size        int          `json:"size,string"`
	Output      AdListOutput `json:"-"`
	doFn        func(ctx context.Context, handler *AdListInput) (out apihandler.OutputI, err error)
}

type AdListOutput struct {
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	Items      []AdListOutputItem `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

type AdListOutputItem struct {
	ID           string `json:"id"`
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

func (o AdListOutput) String() (out string, err error) {
	return apihandler.JsonMarshal(o)
}

// 提供外部设置doFn 入口
func (i *AdListInput) SetDoFn(doFn func(ctx context.Context, handler *AdListInput) (out apihandler.OutputI, err error)) {
	i.doFn = doFn
}

func (i *AdListInput) GetDoFn() (doFn func(ctx context.Context) (out apihandler.OutputI, err error)) {
	return func(ctx context.Context) (out apihandler.OutputI, err error) {
		return i.doFn(ctx, i)
	}
}
func (i *AdListInput) GetRoute() (method string, path string) {
	path = "/api/v1/adList"
	return http.MethodPost, path
}
func (i *AdListInput) GetInputSchema() (lineschema string) {
	lineschema = `
	version=http://json-schema.org/draft-07/schema#,id=in,direction=in
	fullname=title,dst=title,required,description=广告标题,comment=广告标题,example=新年豪礼
	fullname=advertiserId,dst=advertiserId,required,description=广告主,comment=广告主,example=123
	fullname=beginAt,dst=beginAt,required,description=可以投放开始时间,comment=可以投放开始时间,example=2023-01-12 00:00:00
	fullname=endAt,dst=endAt,required,description=投放结束时间,comment=投放结束时间,example=2023-01-30 00:00:00
	fullname=index,dst=index,required,description=页索引,0开始,default=0,comment=页索引,0开始
	fullname=size,dst=size,required,description=每页数量,default=10,comment=每页数量
	fullname=content-type,dst=content-type,required,description=文件格式,default=application/json,comment=文件格式
	fullname=appid,dst=appid,required,description=访问服务的备案id,comment=访问服务的备案id
	fullname=signature,dst=signature,required,description=签名,外网访问需开启签名,comment=签名,外网访问需开启签名
	`
	return
}

func (i *AdListInput) GetOutputSchema() (lineschema string) {
	lineschema = `
	version=http://json-schema.org/draft-07/schema#,id=out,direction=out
	fullname=code,src=code,description=业务状态码,comment=业务状态码,example=0
	fullname=message,src=message,description=业务提示,comment=业务提示,example=ok
	fullname=items,src=items,type=array,description=数组,comment=数组,example=-
	fullname=items[].id,src=items[].id,description=主键,comment=主键,example=0
	fullname=items[].title,src=items[].title,description=广告标题,comment=广告标题,example=新年豪礼
	fullname=items[].advertiserId,src=items[].advertiserId,description=广告主,comment=广告主,example=123
	fullname=items[].summary,src=items[].summary,description=广告素材-文字描述,comment=广告素材-文字描述,example=下单有豪礼
	fullname=items[].image,src=items[].image,description=广告素材-图片地址,comment=广告素材-图片地址
	fullname=items[].link,src=items[].link,description=连接地址,comment=连接地址
	fullname=items[].type,src=items[].type,description=广告素材(类型),text-文字,image-图片,vido-视频,comment=广告素材(类型),text-文字,image-图片,vido-视频,example=image
	fullname=items[].beginAt,src=items[].beginAt,description=投放开始时间,comment=投放开始时间,example=2023-01-12 00:00:00
	fullname=items[].endAt,src=items[].endAt,description=投放结束时间,comment=投放结束时间,example=2023-01-30 00:00:00
	fullname=items[].remark,src=items[].remark,description=备注,comment=备注,example=营养早餐广告
	fullname=items[].valueObj,src=items[].valueObj,description=json扩展,广告的值属性对象,comment=json扩展,广告的值属性对象,example={"tag":"index"}
	fullname=pagination,src=pagination,type=object,description=对象,comment=对象
	fullname=pagination.index,src=pagination.index,description=页索引,0开始,comment=页索引,0开始,example=0
	fullname=pagination.size,src=pagination.size,description=每页数量,comment=每页数量,example=10
	fullname=pagination.total,src=pagination.total,description=总数,comment=总数,example=60
	`
	return
}
