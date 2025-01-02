package resp

type Output struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (o *Output) WithData(data any) *Output {
	o.Data = data
	return o
}

func (o *Output) WithMsg(msg string) *Output {
	o.Msg = msg
	return o
}
