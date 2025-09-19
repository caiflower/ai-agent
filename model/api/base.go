package api

type Request struct {
	RequestId string `header:"X-Request-Id"`
	UserId    string `header:"X-User-Id"`
}
