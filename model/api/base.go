package api

type Request struct {
	RequestID string `header:"X-Request-Id"`
	User      string `header:"X-User-Id"`
}
