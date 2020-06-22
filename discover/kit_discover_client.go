package discover

import (
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"strconv"
	"sync"
)

type kitDiscoverClient struct {
	Host   string //consul host
	Port   int    //consul port
	client consul.Client
	//连接consul的配置
	config *api.Config
	mutex  sync.Mutex
	//服务实例缓存字段
	instanceMap sync.Map
}

func NewKitDiscoverClient(consulHost string, consulPort int) (DiscoveryClient, error) {
	//创建consul.client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(consulPort)
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consul.NewClient(apiClient)
	return &kitDiscoverClient{
		Host:   consulHost,
		Port:   consulPort,
		config: consulConfig,
		client: client,
	}, err
}

//基于kit的consul服务注册
func (consulClient *kitDiscoverClient) Register(serviceName, instanceId, healthCheckUrl string, instanceHost string, instancePort int, meta map[string]string, logger *log.Logger) bool {
	//构建服务实例元数据
	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    serviceName,
		Address: instanceHost,
		Port:    instancePort,
		Meta:    meta,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + instanceHost + ":" + strconv.Itoa(instancePort) + healthCheckUrl,
			Interval:                       "15s",
		},
	}
	//向consul中发送服务注册
	err := consulClient.client.Register(serviceRegistration)
	if err != nil {
		log.Println("Register Service Error")
	}
	log.Println("Register Service Success")
	return true
}

//基于kit的consul服务注销
func (consulClient *kitDiscoverClient) Deregister(instanceId string, logger *log.Logger) bool {
	//构建包含服务实例id的元数据结构体
	serviceRegistrion := &api.AgentServiceRegistration{
		ID: instanceId,
	}
	//向consul发送服务注销
	err := consulClient.client.Deregister(serviceRegistrion)
	if err != nil {
		logger.Println("Deregister Service Error")
		return false
	}
	log.Println("Deregister Service Success")
	return true
}

//基于kit的consul服务发现
func (consulClient *kitDiscoverClient) DiscoverService(serviceName string, logger *log.Logger) []interface{} {
	//查询服务是否已监控并缓存
	instanceList, ok := consulClient.instanceMap.Load(serviceName)
	if ok {
		//直接返回
		return instanceList.([]interface{})
	}
	//申请锁
	consulClient.mutex.Lock()
	//查询服务是否已监控并缓存
	instanceList, ok = consulClient.instanceMap.Load(serviceName)
	if ok {
		return instanceList.([]interface{})
	} else {
		//注册监控
		go func() {
			//使用consul服务实例甲空空某个服务实例列表的变化
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					//数据异常，忽略
					return
				}
				//没有服务实例在线
				if len(v) == 0 {
					consulClient.instanceMap.Store(serviceName, []interface{}{})
				}
				var healthServices []interface{}
				for _, service := range v {
					if service.Checks.AggregatedStatus() == api.HealthPassing {
						healthServices = append(healthServices, service.Service)
					}
				}
				consulClient.instanceMap.Store(serviceName, healthServices)
			}
			defer plan.Stop()
			plan.Run(consulClient.config.Address)

		}()

	}
	defer consulClient.mutex.Unlock()

	//根据服务名 请求服务实例列表
	entries, _, err := consulClient.client.Service(serviceName, "", false, nil)
	if err != nil {
		consulClient.instanceMap.Store(serviceName, []interface{}{})
		logger.Println("Discover Service Error")
		return nil
	}
	instances := make([]interface{}, len(entries))
	for i := 0; i < len(instances); i++ {
		instances[i] = entries[i].Service
	}
	consulClient.instanceMap.Store(serviceName, instances)
	return instances
}
