package ek8s

import (
	"context"
	"fmt"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
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
	logger *elog.Component
	queue  workqueue.Interface
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
		queue:     workqueue.New(),
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

func (c *Component) WatchPrefix(ctx context.Context, appName string, kind string) (err error) {
	for _, ns := range c.config.Namespaces {
		switch kind {
		case KindPod:
			label, err := c.getDeploymentsSelector(ns, appName)
			if err != nil {
				return err
			}
			c.logger.Debug("watch prefix label", zap.String("appname", appName), zap.String("label", label))
			informersFactory := informers.NewSharedInformerFactoryWithOptions(
				c.Clientset,
				defaultResync,
				informers.WithNamespace(ns),
				informers.WithTweakListOptions(func(options *metav1.ListOptions) {
					options.LabelSelector = label
					//// todo
					options.ResourceVersion = "0"
				}),
			)

			informer := informersFactory.Core().V1().Pods()
			c.logger.Debug("k8s watch prefix", zap.String("appname", appName), zap.String("kind", kind), zap.String("kind", kind))
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    c.addPod,
				UpdateFunc: c.updatePod,
				DeleteFunc: c.deletePod,
			})
			// 启动该命名空间里监听
			go informersFactory.Start(ctx.Done())
		case KindEndpoints:
			label, err := c.getServicesSelector(ns, appName)
			if err != nil {
				return err
			}
			c.logger.Debug("watch prefix label", zap.String("appname", appName), zap.String("label", label), zap.String("kind", kind))
			informersFactory := informers.NewSharedInformerFactoryWithOptions(
				c.Clientset,
				defaultResync,
				informers.WithNamespace(ns),
				informers.WithTweakListOptions(func(options *metav1.ListOptions) {
					options.LabelSelector = label
					// todo
					options.ResourceVersion = "0"
				}),
			)

			informer := informersFactory.Core().V1().Endpoints()
			c.logger.Debug("k8s watch prefix", zap.String("appname", appName), zap.String("kind", kind))
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    c.addEndpoints,
				UpdateFunc: c.updateEndpoints,
				DeleteFunc: c.deleteEndpoints,
			})
			// 启动该命名空间里监听
			go informersFactory.Start(ctx.Done())
		default:
			c.logger.Error("k8s watch prefix error", zap.String("appname", appName), zap.String("kind", kind))
		}

	}
	return nil
}

func (c *Component) ProcessWorkItem(f func(info *KubernetesEvent) error) bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)
	o := item.(*KubernetesEvent)
	f(o)
	return true
}

func (c *Component) getDeploymentsSelector(namespace string, appName string) (label string, err error) {
	deployment, err := c.AppsV1().Deployments(namespace).Get(c.getDeploymentName(appName), metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get deployments in namespace (%s), err: %w", namespace, err)
	}
	deploymentLabelMap, err := metav1.LabelSelectorAsMap(deployment.Spec.Selector)
	if err != nil {
		return "", fmt.Errorf("label selector in namespace (%s), err: %w", namespace, err)
	}
	label = labels.SelectorFromSet(deploymentLabelMap).String()
	return
}

func (c *Component) getServicesSelector(namespace string, appName string) (label string, err error) {
	service, err := c.CoreV1().Services(namespace).Get(c.getDeploymentName(appName), metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get services in namespace (%s), err: %w", namespace, err)
	}
	label = labels.SelectorFromSet(service.Labels).String()
	return
}

func (c *Component) getDeploymentName(appName string) string {
	return c.config.DeploymentPrefix + appName
}
