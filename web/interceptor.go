package web

import (
	"github.com/caiflower/ai-agent/constants"
	"github.com/caiflower/common-tools/web"
	"github.com/caiflower/common-tools/web/e"
	"github.com/caiflower/common-tools/web/interceptor"
)

type userInterceptor struct {
}

func NewUserInterceptor() interceptor.Interceptor {
	return &userInterceptor{}
}

func (i *userInterceptor) Before(ctx *web.Context) e.ApiError {
	if ctx.GetAction() != "DescribeHealth" {
		_, r := ctx.GetResponseWriterAndRequest()
		header := r.Header

		if header["X-User-Id"] == nil {
			return e.NewApiError(constants.NotLoginError, "not login", nil)
		}
	}

	return nil
} // 执行业务前执行

func (i *userInterceptor) After(ctx *web.Context, err e.ApiError) e.ApiError {
	return nil
} // 执行业务后执行，参数err为业务返回的ApiErr信息

func (i *userInterceptor) OnPanic(ctx *web.Context, err interface{}) e.ApiError {
	return nil
} // 发生panic时执行
