package plugins

import (
	log2 "github.com/go-kit/kit/log"
	"gomicro-discover/string-service/service"
	"time"
)

//为了记录service接口的执行和调用情况，我们使用装饰者模式定义了loggingMiddleware日志中间件
//使得每当service接口中的方法被调用时都会有对应的日志输出

type loggingMiddleware struct {
	service.Service
	logger log2.Logger
}

func LoggingMiddleware(logger log2.Logger) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return loggingMiddleware{next, logger}
	}
}

func (mw loggingMiddleware) Concat(a, b string) (ret string, err error) {
	//在函数执行结束后打印日志
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Concat",
			"a", a,
			"b", b,
			"result", ret,
			"took", time.Since(begin),
		)
	}(time.Now())
	ret, err = mw.Service.Concat(a, b)
	return ret, err
}

func (mw loggingMiddleware) Diff(a, b string) (ret string, err error) {
	//在函数执行结束后打印日志
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Diff",
			"a", a,
			"b", b,
			"result", ret,
			"took", time.Since(begin),
		)
	}(time.Now())
	ret, err = mw.Service.Diff(a, b)
	return ret, err
}

func (mw loggingMiddleware) HealthCheck() (ret bool) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthCheck",
			"result", ret,
			"took", time.Since(begin),
		)
	}(time.Now())
	ret = mw.Service.HealthCheck()
	return ret
}
