package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"gomicro-discover/service"
)

//endpoint层需要定义返回Endpoint的构建函数，用于将请求转化为Service接口可以处理的参数
//并将处理结果封装成response返回给transport层

//服务发现 Endpoint对象
//endpoint.Endpoint Endpoint是服务器和客户端的基本构建块,它代表单个RPC方法
type DiscoveryEndpoint struct {
	SayHelloEndpoint    endpoint.Endpoint
	DiscoveryEndpoint   endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

//打招呼请求结构体
type SayHelloRequest struct {
}

//打招呼响应体
type SayHelloResponse struct {
	Message string `json:"message"`
}

//创建打招呼Endpoint
func MakeSayHelloEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		message := svc.SayHello()
		return SayHelloResponse{
			Message: message,
		}, nil
	}
}

//服务发现请求结构体
type DiscoveryRequest struct {
	ServiceName string
}

//服务发现响应结构体
type DiscoveryResponse struct {
	Instances []interface{} `json:"instances"`
	Error     string        `json:"error"`
}

//创建服务发现的Endpoint,他是一个rpc类型的函数
func MakeDiscoveryEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		//判断request是否满足DiscoveryRequest
		req := request.(DiscoveryRequest)
		instances, err := svc.DiscoveryService(ctx, req.ServiceName)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}
		return &DiscoveryResponse{
			Instances: instances,
			Error:     errString,
		}, nil
	}
}

//健康检查请求结构体
type HealthRequest struct {
}

//健康检查响应结构体
type HealthResponse struct {
	Status bool `json:"status"`
}

//创建健康检查的创建服务发现的Endpoint
func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return &HealthResponse{
			Status: status,
		}, nil
	}
}
