package endpoint

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"gomicro-discover/string-service/service"
	"strings"
)

//依赖service层提供的方法

var ErrInvalidRequestType = errors.New("RequestType has only two types:concat,diff")

type StringEndpoint struct {
	StringEndpoint   endpoint.Endpoint
	HealCheckEnpoint endpoint.Endpoint
}

type StringRequest struct {
	RequestType string `json:"request_type"`
	A           string `json:"a"`
	B           string `json:"b"`
}

type StringReponse struct {
	Result string `json:"result"`
	Error  error  `json:"error"`
}

func MakeStringEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(StringRequest)
		var (
			res, a, b string
			opError   error
		)
		a = req.A
		b = req.B
		//根据请求操作类型选择操作方法
		if strings.EqualFold(req.RequestType, "Concat") {
			res, _ = svc.Concat(a, b)
		} else if strings.EqualFold(req.RequestType, "Diff") {
			res, _ = svc.Diff(a, b)
		} else {
			return nil, ErrInvalidRequestType
		}
		return StringReponse{Result: res, Error: opError}, nil
	}
}

type HealthRequest struct {
}

type HealthResponse struct {
	Status bool `json:"status"`
}

//创建健康检查的endpoint
func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{
			Status: status,
		}, nil
	}
}
