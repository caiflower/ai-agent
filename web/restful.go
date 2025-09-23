package web

import (
	"net/http"

	. "github.com/caiflower/common-tools/web/v1"
)

func register() {
	Register(NewRestFul().Method(http.MethodGet).Version("v1").Controller("v1.healthController").Path("/healthz").Action("DescribeHealth"))
	Register(NewRestFul().Method(http.MethodGet).Version("v1").Controller("v1.agentController").Path("/chat").Action("Chat"))
}
