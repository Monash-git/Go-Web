package web

//错误格式
type Result struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}