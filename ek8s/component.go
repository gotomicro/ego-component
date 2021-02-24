package ek8s

import (
	"context"
	"fmt"
	"github.com/gotomicro/ego/core/elog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const PackageName = "component.ek8s"
const defaultResync = 5 * time.Minute

// Component ...
type Component struct {
	name   string
	config *Config
	*kubernetes.Clientset
	logger *elog.Component
	queue  workqueue.Interface
}

type KubernetesEvent struct {
	Pod       *v1.Pod
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

func (c *Component) ListPod(appName string) (pods []v1.Pod, err error) {
	pods = make([]v1.Pod, 0)
	for _, ns := range c.config.Namespaces {
		label, err := c.getSelector(ns, appName)
		if err != nil {
			return nil, err
		}
		v1Pods, err := c.CoreV1().Pods(ns).List(metav1.ListOptions{
			LabelSelector: label,
		})

		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods.Items...)
	}
	return
}

func (c *Component) WatchPrefix(ctx context.Context, appName string) (err error) {
	for _, ns := range c.config.Namespaces {
		label, err := c.getSelector(ns, appName)
		if err != nil {
			return err
		}
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
		podInformer := informersFactory.Core().V1().Pods()

		podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.addPod,
			UpdateFunc: c.updatePod,
			DeleteFunc: c.deletePod,
		})
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

func (c *Component) addPod(obj interface{}) {
	p, ok := obj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}
	c.queue.Add(&KubernetesEvent{
		EventType: watch.Added,
		Pod:       p,
	})
}

func (c *Component) updatePod(oldObj, newObj interface{}) {
	op, ok := oldObj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", oldObj)
		return
	}
	np, ok := newObj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", newObj)
		return
	}
	if op.GetResourceVersion() == np.GetResourceVersion() {
		return
	}
	c.queue.Add(&KubernetesEvent{
		Pod:       np,
		EventType: watch.Modified,
	})
}

func (c *Component) deletePod(obj interface{}) {
	p, ok := obj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}
	c.queue.Add(&KubernetesEvent{
		Pod:       p,
		EventType: watch.Deleted,
	})
}

func (c *Component) getSelector(namespace string, appName string) (label string, err error) {
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

func (c *Component) getDeploymentName(appName string) string {
	return c.config.DeploymentPrefix + appName
}
