package eetcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/elog"
	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc"
)

const PackageName = "component.eetcd"

// Component ...
type Component struct {
	name   string
	config *config
	logger *elog.Component
	*clientv3.Client
}

// New ...
func newComponent(name string, config *config, logger *elog.Component) *Component {
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
		grpc.FailOnNonTempDialError(config.EnableFailOnNonTempDialError),
	}

	if config.EnableBlock {
		dialOptions = append(dialOptions, grpc.WithBlock())
	}

	conf := clientv3.Config{
		Endpoints:            config.Addrs,
		DialTimeout:          config.ConnectTimeout,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		DialOptions:          dialOptions,
		AutoSyncInterval:     config.AutoSyncInterval,
	}

	if config.Addrs == nil {
		logger.Panic("client etcd endpoints empty", elog.FieldValueAny(config))
	}

	logger = logger.With(elog.FieldAddr(fmt.Sprintf("%s", config.Addrs)))

	if !config.EnableSecure {
		conf.DialOptions = append(conf.DialOptions, grpc.WithInsecure())
	}

	if config.EnableBasicAuth {
		conf.Username = config.UserName
		conf.Password = config.Password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if config.CaCert != "" {
		certBytes, err := ioutil.ReadFile(config.CaCert)
		if err != nil {
			logger.Panic("parse CaCert failed", elog.Any("err", err))
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if config.CertFile != "" && config.KeyFile != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			logger.Panic("load CertFile or KeyFile failed", elog.Any("config", config), elog.Any("err", err))
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		conf.TLS = tlsConfig
	}

	client, err := clientv3.New(conf)
	if err != nil {
		logger.Panic("client etcd start panic", elog.FieldErr(err), elog.FieldValueAny(config))
	}

	cc := &Component{
		name:   name,
		logger: logger,
		Client: client,
		config: config,
	}

	logger.Info("dial etcd server")
	return cc
}

// GetKeyValue queries etcd key, returns mvccpb.KeyValue
func (c *Component) GetKeyValue(ctx context.Context, key string) (kv *mvccpb.KeyValue, err error) {
	rp, err := c.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(rp.Kvs) > 0 {
		return rp.Kvs[0], nil
	}

	return
}

// GetPrefix get prefix
func (c *Component) GetPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	var (
		vars = make(map[string]string)
	)

	resp, err := c.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return vars, err
	}

	for _, kv := range resp.Kvs {
		vars[string(kv.Key)] = string(kv.Value)
	}

	return vars, nil
}

// DelPrefix 按前缀删除
func (c *Component) DelPrefix(ctx context.Context, prefix string) (deleted int64, err error) {
	resp, err := c.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}
	return resp.Deleted, err
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Component) GetValues(ctx context.Context, keys ...string) (map[string]string, error) {
	var (
		firstRevision = int64(0)
		vars          = make(map[string]string)
		maxTxnOps     = 128
		getOps        = make([]string, 0, maxTxnOps)
	)

	doTxn := func(ops []string) error {
		txnOps := make([]clientv3.Op, 0, maxTxnOps)

		for _, k := range ops {
			txnOps = append(txnOps, clientv3.OpGet(k,
				clientv3.WithPrefix(),
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
				clientv3.WithRev(firstRevision)))
		}

		result, err := c.Txn(ctx).Then(txnOps...).Commit()
		if err != nil {
			return err
		}
		for i, r := range result.Responses {
			originKey := ops[i]
			originKeyFixed := originKey
			if !strings.HasSuffix(originKeyFixed, "/") {
				originKeyFixed = originKey + "/"
			}
			for _, ev := range r.GetResponseRange().Kvs {
				k := string(ev.Key)
				if k == originKey || strings.HasPrefix(k, originKeyFixed) {
					vars[string(ev.Key)] = string(ev.Value)
				}
			}
		}
		if firstRevision == 0 {
			firstRevision = result.Header.GetRevision()
		}
		return nil
	}
	for _, key := range keys {
		getOps = append(getOps, key)
		if len(getOps) >= maxTxnOps {
			if err := doTxn(getOps); err != nil {
				return vars, err
			}
			getOps = getOps[:0]
		}
	}
	if len(getOps) > 0 {
		if err := doTxn(getOps); err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// GetLeaseSession 创建租约会话
func (c *Component) GetLeaseSession(ctx context.Context, opts ...concurrency.SessionOption) (leaseSession *concurrency.Session, err error) {
	return concurrency.NewSession(c.Client, opts...)
}
