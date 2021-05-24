package ek8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotomicro/ego/core/elog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const PackageName = "component.ek8s"
const defaultResync = 5 * time.Minute

const (
	KindPods      = "pods"
	KindEndpoints = "endpoints"
)

// Component ...
type Component struct {
	name   string
	config *Config
	*kubernetes.Clientset
	logger *elog.Component
	locker sync.RWMutex
}

type KubernetesEvent struct {
	IPs       []string
	EventType watch.EventType
}

// New ...
func newComponent(name string, config *Config, logger *elog.Component) *Component {
	client, err := kubernetes.NewForConfig(config.toRestConfig())
	if err != nil {
		logger.Panic("new component err", elog.FieldErr(err))
	}
	return &Component{
		name:      name,
		config:    config,
		logger:    logger,
		Clientset: client,
	}
}

func (c *Component) Config() Config {
	return *c.config
}

func (c *Component) ListPods(appName string) (pods []*v1.Pod, err error) {
	pods = make([]*v1.Pod, 0)
	for _, ns := range c.config.Namespaces {
		v1Pods, err := c.CoreV1().Pods(ns).Get(context.Background(), c.getDeploymentName(appName), metav1.GetOptions{})

		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods)
	}
	return
}

func (c *Component) ListEndpoints(appName string) (pods []*v1.Endpoints, err error) {
	pods = make([]*v1.Endpoints, 0)
	for _, ns := range c.config.Namespaces {
		v1Pods, err := c.CoreV1().Endpoints(ns).Get(context.Background(), c.getDeploymentName(appName), metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods)
	}
	return
}

func (c *Component) NewWatcherApp(ctx context.Context, appName string, kind string) (app *WatcherApp, err error) {
	app = newWatcherApp(c.Clientset, appName, kind, c.config.DeploymentPrefix, c.logger)
	for _, ns := range c.config.Namespaces {
		err = app.watch(ctx, ns)
		if err != nil {
			return app, err
		}
	}
	return app, nil
}

func (c *Component) getDeploymentName(appName string) string {
	return c.config.DeploymentPrefix + appName
}
