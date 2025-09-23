package web

import (
	. "github.com/caiflower/common-tools/web/v1"
)

func StartUp() {
	DefaultHttpServer.AddInterceptor(NewUserInterceptor(), 1)

	// register
	register()

	DefaultHttpServer.StartUp()
}
