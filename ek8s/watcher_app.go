package ek8s

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type WatcherApp struct {
	appName string
	kind    string
	*kubernetes.Clientset
	queue            workqueue.Interface
	logger           *elog.Component
	deploymentPrefix string
}

func newWatcherApp(clientSet *kubernetes.Clientset, appName string, kind string, deploymentPrefix string, logger *elog.Component) *WatcherApp {
	return &WatcherApp{
		Clientset:        clientSet,
		appName:          appName,
		kind:             kind,
		queue:            workqueue.New(),
		deploymentPrefix: deploymentPrefix,
		logger:           logger,
	}
}

func (c *WatcherApp) watch(ctx context.Context, ns string) error {
	switch c.kind {
	case KindPods:
		label, err := c.getDeploymentsSelector(ns, c.appName)
		if err != nil {
			return err
		}
		c.logger.Debug("watch prefix label", zap.String("appname", c.appName), zap.String("label", label))
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
		c.logger.Debug("k8s watch pods", zap.String("appname", c.appName), zap.String("kind", c.kind), zap.String("kind", c.kind))
		informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.addPod,
			UpdateFunc: c.updatePod,
			DeleteFunc: c.deletePod,
		})
		// 启动该命名空间里监听
		go informersFactory.Start(ctx.Done())
	case KindEndpoints:
		endPoints, err := c.CoreV1().Endpoints(ns).Get(context.Background(), c.getDeploymentName(c.appName), metav1.GetOptions{})
		if err != nil {
			return err
		}

		c.logger.Info("k8s watch endpoints", zap.String("appname", c.appName), zap.String("namespace", endPoints.Namespace), zap.String("endPointName", endPoints.Name), zap.String("kind", c.kind))
		informersFactory := informers.NewSharedInformerFactoryWithOptions(
			c.Clientset,
			defaultResync,
			informers.WithNamespace(ns),
			informers.WithTweakListOptions(func(options *metav1.ListOptions) {
				options.FieldSelector = "metadata.name=" + endPoints.Name
				// todo
				options.ResourceVersion = "0"
			}),
		)

		informer := informersFactory.Core().V1().Endpoints()
		informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.addEndpoints,
			UpdateFunc: c.updateEndpoints,
			DeleteFunc: c.deleteEndpoints,
		})
		// 启动该命名空间里监听
		go informersFactory.Start(ctx.Done())
	default:
		c.logger.Error("k8s watch prefix error", zap.String("appname", c.appName), zap.String("kind", c.kind))
	}
	return nil
}

func (c *WatcherApp) ProcessWorkItem(f func(info *KubernetesEvent) error) bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)
	o := item.(*KubernetesEvent)
	f(o)
	return true
}

func (c *WatcherApp) getDeploymentsSelector(namespace string, appName string) (label string, err error) {
	deployment, err := c.AppsV1().Deployments(namespace).Get(context.Background(), c.getDeploymentName(appName), metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get deployments in namespace (%s), err: %w", namespace, err)
	}
	label = labels.SelectorFromSet(deployment.Labels).String()
	return
}

func (c *WatcherApp) getServicesSelector(namespace string, appName string) (label string, err error) {
	service, err := c.CoreV1().Services(namespace).Get(context.Background(), c.getDeploymentName(appName), metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get services in namespace (%s), err: %w", namespace, err)
	}
	label = labels.SelectorFromSet(service.Labels).String()
	return
}

func (c *WatcherApp) getDeploymentName(appName string) string {
	return c.deploymentPrefix + appName
}
