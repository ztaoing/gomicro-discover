package main

import (
	"context"
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"gomicro-discover/discover"
	"gomicro-discover/string-service/config"
	"gomicro-discover/string-service/endpoint"
	"gomicro-discover/string-service/plugins"
	"gomicro-discover/string-service/service"
	"gomicro-discover/string-service/transport"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {

	var (
		servicePort = flag.Int("service.port", 10085, "service port")
		serviceHost = flag.String("service.host", "127.0.0.1", "service host")

		consulPort = flag.Int("consul.port", 8500, "consul port")
		consulHost = flag.String("consul.host", "127.0.0.1", "consul host")

		serviceName = flag.String("service.name", "string", "service name")
	)
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	var discoveryClient discover.DiscoveryClient
	discoveryClient, err := discover.NewKitDiscoverClient(*consulHost, *consulPort)
	if err != nil {
		config.Logger.Println("Get Consul Client failed")
		os.Exit(-1)
	}

	var svc service.Service
	svc = service.StringService{}

	svc = plugins.LoggingMiddleware(config.KitLogger)(svc)

	stringEndpoint := endpoint.MakeStringEndpoint(svc)

	//创建健康检查Endpoint
	healthEndpoint := endpoint.MakeHealthCheckEndpoint(svc)

	//封装到StringEndpoints
	endpts := endpoint.StringEndpoint{
		StringEndpoint:      stringEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpts, config.KitLogger)
	instanceId := *serviceName + "-" + uuid.NewV4().String()

	//http server
	go func() {
		config.Logger.Println("string-service for service %s failed", serviceName)
		//注册服务
		if !discoveryClient.Register(*serviceName, instanceId, "/health", *serviceHost, *servicePort, nil, config.Logger) {
			config.Logger.Println("string-service for service %s failed", serviceName)
			os.Exit(-1)
		}
		handler := r
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)

	}()

	error := <-errChan
	//注销服务
	discoveryClient.Deregister(instanceId, config.Logger)
	config.Logger.Println(error)
}
