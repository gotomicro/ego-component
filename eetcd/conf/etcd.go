package conf

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/econf/manager"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/util/xcast"
	"github.com/gotomicro/ego/core/util/xgo"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"

	"github.com/gotomicro/ego-component/eetcd"
)

// dataSource file provider.
type dataSource struct {
	key                 string
	enableWatch         bool
	lastUpdatedRevision int64
	changed             chan struct{}
	cancel              context.CancelFunc
	logger              *elog.Component
	etcd                *eetcd.Component
}

func init() {
	manager.Register("etcd", &dataSource{})
}

// Parse
func (fp *dataSource) Parse(path string, watch bool) econf.ConfigType {
	fp.logger = elog.EgoLogger.With(elog.FieldComponent(econf.PackageName))

	urlInfo, err := url.Parse(path)
	if err != nil {
		fp.logger.Panic("new datasource", elog.FieldErr(err))
		return ""
	}

	configKey := urlInfo.Query().Get("configKey")
	configType := urlInfo.Query().Get("configType")

	if configKey == "" {
		fp.logger.Panic("key is empty")
	}

	if configType == "" {
		fp.logger.Panic("configType is empty")
	}

	fp.etcd = eetcd.DefaultContainer().Build(
		eetcd.WithAddrs([]string{urlInfo.Host}),
		eetcd.WithEnableBasicAuth(xcast.ToBool(urlInfo.Query().Get("basicAuth"))),
		eetcd.WithEnableSecure(xcast.ToBool(urlInfo.Query().Get("secure"))),
		eetcd.WithCertFile(urlInfo.Query().Get("certFile")),
		eetcd.WithKeyFile(urlInfo.Query().Get("keyFile")),
		eetcd.WithCaCert(urlInfo.Query().Get("caCert")),
		eetcd.WithUserName(urlInfo.Query().Get("username")),
		eetcd.WithPassword(urlInfo.Query().Get("password")),
	)

	fp.key = configKey
	fp.enableWatch = watch

	if watch {
		fp.changed = make(chan struct{}, 1)
		xgo.Go(fp.watch)
	}
	return econf.ConfigType(configType)
}

// ReadConfig ...
func (fp *dataSource) ReadConfig() (content []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := fp.etcd.Get(ctx, fp.key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, errors.New("empty response")
	}
	fp.lastUpdatedRevision = resp.Header.GetRevision()
	return resp.Kvs[0].Value, nil
}

func (fp *dataSource) handle(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > fp.lastUpdatedRevision {
		fp.lastUpdatedRevision = resp.CompactRevision
	}
	if resp.Header.GetRevision() > fp.lastUpdatedRevision {
		fp.lastUpdatedRevision = resp.Header.GetRevision()
	}

	if err := resp.Err(); err != nil {
		return
	}

	for _, ev := range resp.Events {
		if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
			select {
			case fp.changed <- struct{}{}:
			default:
			}
		}
	}
}

// Close ...
func (fp *dataSource) Close() error {
	fp.cancel()
	return nil
}

// IsConfigChanged ...
func (fp *dataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

// Watch file and automate update.
func (fp *dataSource) watch() {
	ctx, cancel := context.WithCancel(context.Background())
	fp.cancel = cancel
	rch := fp.etcd.Watch(ctx, fp.key, clientv3.WithCreatedNotify(), clientv3.WithRev(fp.lastUpdatedRevision))
	for {
		for resp := range rch {
			fp.handle(&resp)
		}
		time.Sleep(time.Second)

		ctx, cancel = context.WithCancel(context.Background())
		if fp.lastUpdatedRevision > 0 {
			rch = fp.etcd.Watch(ctx, fp.key, clientv3.WithCreatedNotify(), clientv3.WithRev(fp.lastUpdatedRevision))
		} else {
			rch = fp.etcd.Watch(ctx, fp.key, clientv3.WithCreatedNotify())
		}
		fp.cancel = cancel
	}
}
