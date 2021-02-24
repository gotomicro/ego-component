package registry

import (
	"context"
	"fmt"
	"github.com/gotomicro/ego-component/ekubernetes"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/eregistry"
	"github.com/gotomicro/ego/server"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"sync"
)

type Component struct {
	name   string
	client *ek8s.Component
	kvs    sync.Map
	Config *Config
	cancel context.CancelFunc
	rmu    *sync.RWMutex
	logger *elog.Component
}

func newComponent(name string, config *Config, logger *elog.Component, client *ek8s.Component) *Component {
	reg := &Component{
		name:   name,
		logger: logger,
		client: client,
		Config: config,
		kvs:    sync.Map{},
		rmu:    &sync.RWMutex{},
	}
	return reg
}

// RegisterService register service to registry
func (reg *Component) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	return nil
}

// UnregisterService unregister service from registry
func (reg *Component) UnregisterService(ctx context.Context, info *server.ServiceInfo) error {
	return nil
}

// ListServices list service registered in registry with name `name`
func (reg *Component) ListServices(ctx context.Context, addr string, scheme string) (services []*server.ServiceInfo, err error) {
	appName, port, err := getAppnameAndPort(addr)
	if err != nil {
		return nil, err
	}

	getResp, getErr := reg.client.ListPod(appName)
	if getErr != nil {
		reg.logger.Error("watch request err", elog.FieldErrKind("request err"), elog.FieldErr(getErr), elog.FieldAddr(appName))
		return nil, getErr
	}

	for _, kv := range getResp {
		var service server.ServiceInfo
		service.Address = kv.Status.PodIP + ":" + port
		services = append(services, &service)
	}
	return
}

// WatchServices watch service change event, then return address list
func (reg *Component) WatchServices(ctx context.Context, addr string, scheme string) (chan eregistry.Endpoints, error) {
	appName, port, err := getAppnameAndPort(addr)
	if err != nil {
		return nil, err
	}

	err = reg.client.WatchPrefix(ctx, appName)
	if err != nil {
		return nil, err
	}

	svcs, err := reg.ListServices(ctx, addr, scheme)
	if err != nil {
		return nil, err
	}
	var al = &eregistry.Endpoints{
		Nodes:           make(map[string]server.ServiceInfo),
		RouteConfigs:    make(map[string]eregistry.RouteConfig),
		ConsumerConfigs: make(map[string]eregistry.ConsumerConfig),
		ProviderConfigs: make(map[string]eregistry.ProviderConfig),
	}
	var addresses = make(chan eregistry.Endpoints, 10)

	for _, svc := range svcs {
		reg.updateAddrList(al, svc.Address)
	}

	addresses <- *al.DeepCopy()
	go func() {

		for reg.client.ProcessWorkItem(func(info *ek8s.KubernetesEvent) error {
			switch info.EventType {
			case watch.Added:
				reg.updateAddrList(al, info.Pod.Status.PodIP+":"+port)
			case watch.Deleted:
				reg.deleteAddrList(al, info.Pod.Status.PodIP+":"+port)
			}

			out := al.DeepCopy()
			select {
			case addresses <- *out:
			default:
				elog.Warnf("invalid")
			}
			return nil
		}) {
		}
	}()

	return addresses, nil
}

// Close ...
func (reg *Component) Close() error {
	if reg.cancel != nil {
		reg.cancel()
	}
	return nil
}

func (reg *Component) deleteAddrList(al *eregistry.Endpoints, addr string) {
	delete(al.Nodes, addr)
}

func (reg *Component) updateAddrList(al *eregistry.Endpoints, addr string) {
	al.Nodes[addr] = server.ServiceInfo{
		Address: addr,
	}
}

func getAppnameAndPort(addr string) (appName, port string, err error) {
	if !strings.Contains(addr, ":") {
		err = fmt.Errorf("getAppnameAndPort addr is %s, and must have `:` and `port`", addr)
		return
	}
	arrs := strings.Split(addr, ":")
	if len(arrs) != 2 {
		err = fmt.Errorf("getAppnameAndPort length error")
		return
	}
	appName = arrs[0]
	port = arrs[1]
	return
}
