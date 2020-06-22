package service

import (
	"context"
	"errors"
	"gomicro-discover/config"
	"gomicro-discover/discover"
)

//服务接口

type Service interface {
	//健康检查接口
	HealthCheck() bool
	//打招呼接口
	SayHello() string
	//服务发现接口
	DiscoveryService(ctx context.Context, serviceName string) ([]interface{}, error)
}

var errNotServiceInstance = errors.New("instances are not existed")

type DiscoveryServiceImpl struct {
	discoverClient discover.DiscoveryClient
}

//返回Service接口，DiscoveryServiceImpl必须实现了Service接口
func NewDiscoverServiceImpl(discoverClient discover.DiscoveryClient) Service {
	return &DiscoveryServiceImpl{
		discoverClient: discoverClient,
	}
}

func (service *DiscoveryServiceImpl) SayHello() string {
	return "hello world ha!"
}

func (service *DiscoveryServiceImpl) DiscoveryService(ctx context.Context, serviceName string) ([]interface{}, error) {
	instances := service.discoverClient.DiscoverService(serviceName, config.Logger)
	if instances == nil || len(instances) == 0 {
		return nil, errNotServiceInstance
	}
	return instances, nil
}

//用于检测服务的健康状态，这里不做处理，直接返回true
func (service *DiscoveryServiceImpl) HealthCheck() bool {
	return true
}
