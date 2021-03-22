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
	KindPod       = "Pod"
	KindEndpoints = "Endpoints"
)

// Component ...
type Component struct {
	name   string
	config *Config
	*kubernetes.Clientset
	logger   *elog.Component
	watchApp map[string]*WatcherApp
	locker   sync.RWMutex
	//queue  workqueue.Interface
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
		watchApp:  make(map[string]*WatcherApp),
	}
}

func (c *Component) ListPod(appName string) (pods []*v1.Pod, err error) {
	pods = make([]*v1.Pod, 0)
	for _, ns := range c.config.Namespaces {
		v1Pods, err := c.CoreV1().Pods(ns).Get(c.getDeploymentName(appName), metav1.GetOptions{})

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
		v1Pods, err := c.CoreV1().Endpoints(ns).Get(c.getDeploymentName(appName), metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods)
	}
	return
}

func (c *Component) NewWatcherApp(ctx context.Context, appName string, kind string) (app *WatcherApp, err error) {
	var flag bool
	c.locker.RLock()
	app, flag = c.watchApp[appName]
	if flag {
		c.locker.RUnlock()
		return app, nil
	}
	c.locker.RUnlock()

	app = newWatcherApp(c.Clientset, appName, kind, c.config.DeploymentPrefix, c.logger)
	for _, ns := range c.config.Namespaces {
		err = app.watch(ctx, ns)
		if err != nil {
			return app, err
		}
	}

	c.locker.Lock()
	c.watchApp[appName] = app
	c.locker.Unlock()
	return app, nil
}

func (c *Component) getDeploymentName(appName string) string {
	return c.config.DeploymentPrefix + appName
}
