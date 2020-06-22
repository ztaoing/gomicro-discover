package main

import (
	"context"
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"gomicro-discover/config"
	"gomicro-discover/discover"
	"gomicro-discover/endpoint"
	"gomicro-discover/service"
	"gomicro-discover/transport"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

//从命令行中读取相关参数，没有时，使用默认值

func main() {
	var (
		//服务地址、端口、服务名
		servicePort = flag.Int("service.port", 10086, "service port")
		serviceHost = flag.String("service.host", "127.0.0.1", "service host")
		serviceName = flag.String("service.name", "SayHello", "service name")

		//consul 地址
		consulPort = flag.Int("consul.port", 8500, "consul port")
		consulHost = flag.String("consul.host", "127.0.0.1", "consul host")
	)
	flag.Parse()
	ctx := context.Background()
	errChan := make(chan error)

	//生命服务发现客户端
	var discoverClient discover.DiscoveryClient

	discoverClient, err := discover.NewKitDiscoverClient(*consulHost, *consulPort)

	//获取服务发现客户端失败，直接关闭服务
	if err != nil {
		config.Logger.Println("Get consul Client failed")
		os.Exit(-1)
	}

	//声明并初始化service
	var svc = service.NewDiscoverServiceImpl(discoverClient)

	//创建endpoint
	sayHellopoint := endpoint.MakeSayHelloEndpoint(svc)
	discoveryEndpoint := endpoint.MakeDiscoveryEndpoint(svc)
	healthEndpoint := endpoint.MakeHealthCheckEndpoint(svc)

	endpts := endpoint.DiscoveryEndpoint{
		SayHelloEndpoint:    sayHellopoint,
		DiscoveryEndpoint:   discoveryEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	//创建http.handler
	r := transport.MakeHttpHandler(ctx, endpts, config.KitLogger)
	//定义服务实例id
	instanceId := *serviceName + "-" + uuid.NewV4().String()

	//启动httpserver
	go func() {
		config.Logger.Println("Http Server start at port:" + strconv.Itoa(*servicePort))

		//注册服务
		if !discoverClient.Register(*serviceName, instanceId, "/health", *serviceHost, *servicePort, nil, config.Logger) {
			//注册失败
			config.Logger.Printf("string-service for service %s failed.", serviceName)
			os.Exit(-1)
		}
		handler := r
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), handler)
	}()

	//监控系统信号
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	//服务退出取消注册
	discoverClient.Deregister(instanceId, config.Logger)
	config.Logger.Println(error)
}
