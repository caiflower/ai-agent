package constants

import (
	"net/http"

	"github.com/caiflower/common-tools/web/e"
)

var NotLoginError = &e.ErrorCode{Code: http.StatusUnauthorized, Type: "NotLogin"}
