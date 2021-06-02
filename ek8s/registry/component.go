package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gotomicro/ego/client/egrpc/resolver"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/eregistry"
	"github.com/gotomicro/ego/server"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/gotomicro/ego-component/ek8s"
)

var _ eregistry.Registry = &Component{}

type Component struct {
	name             string
	client           *ek8s.Component
	kvs              sync.Map
	config           *Config
	cancel           context.CancelFunc
	rmu              *sync.RWMutex
	logger           *elog.Component
	fallbackRegistry eregistry.Registry
}

func newComponent(name string, config *Config, logger *elog.Component, client *ek8s.Component) *Component {
	var reg *Component
	reg = &Component{
		name:             name,
		logger:           logger,
		client:           client,
		config:           config,
		kvs:              sync.Map{},
		rmu:              &sync.RWMutex{},
		fallbackRegistry: nil,
	}
	// 注册为grpc的resolver
	resolver.Register(config.Scheme, reg)
	return reg
}

// RegisterService do noting
func (reg *Component) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	return nil
}

// UnregisterService do noting
func (reg *Component) UnregisterService(ctx context.Context, info *server.ServiceInfo) error {
	return nil
}

// ListServices list service registered in registry with name `name`
func (reg *Component) ListServices(ctx context.Context, t eregistry.Target) (services []*server.ServiceInfo, err error) {
	appName, port, err := getAppnameAndPort(t.Endpoint)
	if err != nil {
		return nil, err
	}

	switch reg.config.Kind {
	case ek8s.KindPods:
		getResp, getErr := reg.client.ListPodsByName(appName)
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
	case ek8s.KindEndpoints:
		getResp, getErr := reg.client.ListEndpointsByName(appName)
		if getErr != nil {
			reg.logger.Error("watch request err", elog.FieldErrKind("request err"), elog.FieldErr(getErr), elog.FieldAddr(appName))
			return nil, getErr
		}
		for _, kv := range getResp {
			for _, subsets := range kv.Subsets {
				for _, address := range subsets.Addresses {
					var service server.ServiceInfo
					service.Address = address.IP + ":" + port
					services = append(services, &service)
				}
			}
		}
		elog.Debug("ListServices", zap.Any("services", services))
		return
	default:
		elog.Error("list services error", zap.String("kind", reg.config.Kind))
	}
	return
}

// WatchServices watch service change event, then return address list
func (reg *Component) WatchServices(ctx context.Context, t eregistry.Target) (chan eregistry.Endpoints, error) {
	appName, port, err := getAppnameAndPort(t.Endpoint)
	if err != nil {
		return nil, err
	}

	app, err := reg.client.NewWatcherApp(ctx, appName, reg.config.Kind)
	if err != nil {
		return nil, err
	}

	svcs, err := reg.ListServices(ctx, t)
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
		reg.addAddrList(al, []string{svc.Address})
	}

	addresses <- *al.DeepCopy()
	go func() {
		for app.ProcessWorkItem(func(info *ek8s.KubernetesEvent) error {
			switch info.EventType {
			case watch.Added:
				addrs := make([]string, 0)
				for _, ip := range info.IPs {
					addrs = append(addrs, ip+":"+port)
				}
				reg.addAddrList(al, addrs)
			case watch.Deleted:
				addrs := make([]string, 0)
				for _, ip := range info.IPs {
					addrs = append(addrs, ip+":"+port)
				}
				reg.deleteAddrList(al, addrs)
			case watch.Modified:
				addrs := make([]string, 0)
				for _, ip := range info.IPs {
					addrs = append(addrs, ip+":"+port)
				}
				reg.updateAddrList(al, addrs)
			}
			out := al.DeepCopy()
			reg.logger.Info("update addresses", zap.String("appName", appName), zap.Any("addresses", *out))
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

func (reg *Component) SyncServices(context.Context, eregistry.SyncServicesOptions) error {
	return nil
}

// Close ...
func (reg *Component) Close() error {
	if reg.cancel != nil {
		reg.cancel()
	}
	return nil
}

// IsValid 判断k8s Registry是否有效
func (reg *Component) IsValid(ctx context.Context) error {
	if reg.config.Kind == ek8s.KindPods {
		for _, ns := range reg.client.Config().Namespaces {
			if _, err := reg.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{}); err != nil {
				return fmt.Errorf("k8s component list pod fail, %w", err)
			}
			if _, err := reg.client.CoreV1().Pods(ns).Watch(ctx, metav1.ListOptions{}); err != nil {
				return fmt.Errorf("k8s component watch pod fail, %w", err)
			}
		}
	} else if reg.config.Kind == ek8s.KindEndpoints {
		for _, ns := range reg.client.Config().Namespaces {
			if _, err := reg.client.CoreV1().Endpoints(ns).List(ctx, metav1.ListOptions{}); err != nil {
				return fmt.Errorf("k8s component list endpoints fail, %w", err)
			}
			if _, err := reg.client.CoreV1().Endpoints(ns).Watch(ctx, metav1.ListOptions{}); err != nil {
				return fmt.Errorf("k8s component watch endpoints fail, %w", err)
			}
		}
	}

	return nil
}

func (reg *Component) deleteAddrList(al *eregistry.Endpoints, addrs []string) {
	for _, addr := range addrs {
		delete(al.Nodes, addr)
	}
}

func (reg *Component) addAddrList(al *eregistry.Endpoints, addrs []string) {
	for _, addr := range addrs {
		al.Nodes[addr] = server.ServiceInfo{
			Address: addr,
		}
	}
}

func (reg *Component) updateAddrList(al *eregistry.Endpoints, addrs []string) {
	al.Nodes = make(map[string]server.ServiceInfo)
	for _, addr := range addrs {
		al.Nodes[addr] = server.ServiceInfo{
			Address: addr,
		}
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
