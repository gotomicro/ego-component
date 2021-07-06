package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gotomicro/ego/client/egrpc/resolver"
	"github.com/gotomicro/ego/core/constant"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/eregistry"
	"github.com/gotomicro/ego/core/util/xgo"
	"github.com/gotomicro/ego/server"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/gotomicro/ego-component/eetcd"
)

var _ eregistry.Registry = &Component{}

type Component struct {
	name     string
	client   *eetcd.Component
	kvs      sync.Map
	Config   *Config
	cancel   context.CancelFunc
	rmu      *sync.RWMutex
	sessions map[string]*concurrency.Session
	logger   *elog.Component
}

func newComponent(name string, config *Config, logger *elog.Component, client *eetcd.Component) *Component {
	reg := &Component{
		name:     name,
		logger:   logger,
		client:   client,
		Config:   config,
		kvs:      sync.Map{},
		rmu:      &sync.RWMutex{},
		sessions: make(map[string]*concurrency.Session),
	}
	resolver.Register(config.Scheme, reg)
	return reg
}

// RegisterService register service to registry
func (reg *Component) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	err := reg.registerBiz(ctx, info)
	if err != nil {
		return err
	}
	return reg.registerMetric(ctx, info)
}

// UnregisterService unregister service from registry
func (reg *Component) UnregisterService(ctx context.Context, info *server.ServiceInfo) error {
	return reg.unregister(ctx, reg.registerKey(info))
}

// ListServices list service registered in registry with name `name`
func (reg *Component) ListServices(ctx context.Context, t eregistry.Target) (services []*server.ServiceInfo, err error) {
	key := fmt.Sprintf("/%s/%s/providers/%s://", reg.Config.Prefix, t.Endpoint, t.Protocol)
	getResp, getErr := reg.client.Get(ctx, key, clientv3.WithPrefix())
	if getErr != nil {
		reg.logger.Error("watch request err", elog.FieldErrKind("request err"), elog.FieldErr(getErr), elog.FieldAddr(t.Endpoint))
		return nil, getErr
	}

	for _, kv := range getResp.Kvs {
		var service server.ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			reg.logger.Warnf("invalid service", elog.FieldErr(err))
			continue
		}
		services = append(services, &service)
	}

	return
}

// WatchServices watch service change event, then return address list
func (reg *Component) WatchServices(ctx context.Context, t eregistry.Target) (chan eregistry.Endpoints, error) {
	prefix := fmt.Sprintf("/%s/%s/", reg.Config.Prefix, t.Endpoint)
	watch, err := reg.client.WatchPrefix(context.Background(), prefix)
	if err != nil {
		return nil, err
	}

	var addresses = make(chan eregistry.Endpoints, 10)
	var al = &eregistry.Endpoints{
		Nodes:           make(map[string]server.ServiceInfo),
		RouteConfigs:    make(map[string]eregistry.RouteConfig),
		ConsumerConfigs: make(map[string]eregistry.ConsumerConfig),
		ProviderConfigs: make(map[string]eregistry.ProviderConfig),
	}

	for _, kv := range watch.IncipientKeyValues() {
		reg.updateAddrList(al, prefix, t.Protocol, kv)
	}

	addresses <- *al.DeepCopy()
	xgo.Go(func() {
		for event := range watch.C() {
			switch event.Type {
			case mvccpb.PUT:
				reg.updateAddrList(al, prefix, t.Protocol, event.Kv)
			case mvccpb.DELETE:
				reg.deleteAddrList(al, prefix, t.Protocol, event.Kv)
			}

			out := al.DeepCopy()
			select {
			case addresses <- *out:
			default:
				elog.Warnf("invalid")
			}
		}
	})

	return addresses, nil
}

func (reg *Component) SyncServices(context.Context, eregistry.SyncServicesOptions) error {
	return nil
}

func (reg *Component) unregister(ctx context.Context, key string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reg.Config.ReadTimeout)
		defer cancel()
	}

	if err := reg.delSession(key); err != nil {
		return err
	}

	_, err := reg.client.Delete(ctx, key)
	if err == nil {
		reg.kvs.Delete(key)
	}
	return err
}

// Close ...
func (reg *Component) Close() error {
	if reg.cancel != nil {
		reg.cancel()
	}
	var wg sync.WaitGroup
	reg.kvs.Range(func(k, v interface{}) bool {
		wg.Add(1)
		go func(k interface{}) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := reg.unregister(ctx, k.(string))
			if err != nil {
				reg.logger.Error("unregister service", elog.FieldErrKind("request err"), elog.FieldErr(err), elog.FieldErr(err), elog.FieldKey(fmt.Sprintf("%v", k)), elog.FieldValueAny(v))
			} else {
				reg.logger.Info("unregister service", elog.FieldKey(fmt.Sprintf("%v", k)), elog.FieldValueAny(v))
			}
			cancel()
		}(k)
		return true
	})
	wg.Wait()
	return nil
}

func (reg *Component) registerMetric(ctx context.Context, info *server.ServiceInfo) error {
	if info.Kind != constant.ServiceGovernor {
		return nil
	}

	metric := "/prometheus/job/%s/%s"

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reg.Config.ReadTimeout)
		defer cancel()
	}

	val := info.Address
	key := fmt.Sprintf(metric, info.Name, val)

	opOptions := make([]clientv3.OpOption, 0)
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		//todo ctx without timeout for same as service life?
		sess, err := reg.getSession(key, concurrency.WithTTL(int(ttl)))
		if err != nil {
			return err
		}
		opOptions = append(opOptions, clientv3.WithLease(sess.Lease()))
	}
	_, err := reg.client.Put(ctx, key, val, opOptions...)
	if err != nil {
		reg.logger.Error("register service", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(info))
		return err
	}

	reg.logger.Info("register service", elog.FieldKey(key), elog.FieldValueAny(val))
	reg.kvs.Store(key, val)
	return nil

}
func (reg *Component) registerBiz(ctx context.Context, info *server.ServiceInfo) error {
	var readCtx context.Context
	var readCancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		readCtx, readCancel = context.WithTimeout(ctx, reg.Config.ReadTimeout)
		defer readCancel()
	}

	key := reg.registerKey(info)
	val := reg.registerValue(info)

	opOptions := make([]clientv3.OpOption, 0)
	// opOptions = append(opOptions, clientv3.WithSerializable())
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		//todo ctx without timeout for same as service life?
		sess, err := reg.getSession(key, concurrency.WithTTL(int(ttl)))
		if err != nil {
			return err
		}

		opOptions = append(opOptions, clientv3.WithLease(sess.Lease()))
	}
	_, err := reg.client.Put(readCtx, key, val, opOptions...)
	if err != nil {
		reg.logger.Error("register service", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(info))
		return err
	}
	reg.logger.Info("register service", elog.FieldKey(key), elog.FieldValueAny(val))
	reg.kvs.Store(key, val)
	return nil
}

func (reg *Component) getSession(k string, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	reg.rmu.RLock()
	sess, ok := reg.sessions[k]
	reg.rmu.RUnlock()
	if ok {
		return sess, nil
	}
	sess, err := concurrency.NewSession(reg.client.Client, opts...)
	if err != nil {
		return sess, err
	}
	reg.rmu.Lock()
	reg.sessions[k] = sess
	reg.rmu.Unlock()
	return sess, nil
}

func (reg *Component) delSession(k string) error {
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		reg.rmu.RLock()
		sess, ok := reg.sessions[k]
		reg.rmu.RUnlock()
		if ok {
			reg.rmu.Lock()
			delete(reg.sessions, k)
			reg.rmu.Unlock()
			if err := sess.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (reg *Component) registerKey(info *server.ServiceInfo) string {
	return eregistry.GetServiceKey(reg.Config.Prefix, info)
}

func (reg *Component) registerValue(info *server.ServiceInfo) string {
	return eregistry.GetServiceValue(info)
}

func (reg *Component) deleteAddrList(al *eregistry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)
		if strings.HasPrefix(addr, "providers/"+scheme) {
			// 解析服务注册键
			addr = strings.TrimPrefix(addr, "providers/")
			if addr == "" {
				continue
			}
			uri, err := url.Parse(addr)
			if err != nil {
				reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
				continue
			}
			delete(al.Nodes, uri.String())
		}

		if strings.HasPrefix(addr, "configurators/"+scheme) {
			// 解析服务配置键
			addr = strings.TrimPrefix(addr, "configurators/")
			if addr == "" {
				continue
			}
			uri, err := url.Parse(addr)
			if err != nil {
				reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
				continue
			}
			delete(al.RouteConfigs, uri.String())
		}

		if isIPPort(addr) {
			// 直接删除addr 因为Delete操作的value值为空
			delete(al.Nodes, addr)
			delete(al.RouteConfigs, addr)
		}
	}
}

func (reg *Component) updateAddrList(al *eregistry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)
		switch {
		// 解析服务注册键
		case strings.HasPrefix(addr, "providers/"+scheme):
			addr = strings.TrimPrefix(addr, "providers/")
			uri, err := url.Parse(addr)
			if err != nil {
				reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
				continue
			}
			var serviceInfo server.ServiceInfo
			if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
				reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
				continue
			}
			al.Nodes[uri.String()] = serviceInfo
		case strings.HasPrefix(addr, "configurators/"+scheme):
			addr = strings.TrimPrefix(addr, "configurators/")

			uri, err := url.Parse(addr)
			if err != nil {
				reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
				continue
			}

			if strings.HasPrefix(uri.Path, "/routes/") { // 路由配置
				var routeConfig eregistry.RouteConfig
				if err := json.Unmarshal(kv.Value, &routeConfig); err != nil {
					reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
					continue
				}
				routeConfig.ID = strings.TrimPrefix(uri.Path, "/routes/")
				routeConfig.Scheme = uri.Scheme
				routeConfig.Host = uri.Host
				al.RouteConfigs[uri.String()] = routeConfig
			}

			if strings.HasPrefix(uri.Path, "/providers/") {
				var providerConfig eregistry.ProviderConfig
				if err := json.Unmarshal(kv.Value, &providerConfig); err != nil {
					reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
					continue
				}
				providerConfig.ID = strings.TrimPrefix(uri.Path, "/providers/")
				providerConfig.Scheme = uri.Scheme
				providerConfig.Host = uri.Host
				al.ProviderConfigs[uri.String()] = providerConfig
			}

			if strings.HasPrefix(uri.Path, "/consumers/") {
				var consumerConfig eregistry.ConsumerConfig
				if err := json.Unmarshal(kv.Value, &consumerConfig); err != nil {
					reg.logger.Error("parse uri", elog.FieldErr(err), elog.FieldKey(string(kv.Key)))
					continue
				}
				consumerConfig.ID = strings.TrimPrefix(uri.Path, "/consumers/")
				consumerConfig.Scheme = uri.Scheme
				consumerConfig.Host = uri.Host
				al.ConsumerConfigs[uri.String()] = consumerConfig
			}
		}
	}
}

func isIPPort(addr string) bool {
	_, _, err := net.SplitHostPort(addr)
	return err == nil
}
