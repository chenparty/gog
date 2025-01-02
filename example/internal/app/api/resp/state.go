package resp

type State string

const (
	OK              State = "OK"
	InvalidErr      State = "InvalidParamsErr"
	UnauthorizedErr State = "UnauthorizedErr"
	DBErr           State = "DBErr"
	ServiceErr      State = "ServiceErr"
)

func (r State) text() string {
	switch r {
	case OK:
		return "OK"
	case InvalidErr:
		return "参数异常"
	case UnauthorizedErr:
		return "未授权"
	case DBErr:
		return "数据库异常"
	case ServiceErr:
		return "业务服务异常"
	default:
		return "未知错误"
	}
}

func (r State) Output() *Output {
	return &Output{
		Code: string(r),
		Msg:  r.text(),
		Data: nil,
	}
}
