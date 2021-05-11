package ek8s

import (
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *WatcherApp) addPod(obj interface{}) {
	c.logger.Debug("addPod", zap.Any("obj", obj))
	p, ok := obj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}
	c.queue.Add(&KubernetesEvent{
		EventType: watch.Added,
		IPs:       []string{p.Status.PodIP},
	})
}

func (c *WatcherApp) updatePod(oldObj, newObj interface{}) {
	c.logger.Debug("updatePod", zap.Any("oldObj", oldObj), zap.Any("newObj", newObj))

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
		IPs:       []string{np.Status.PodIP},
		EventType: watch.Modified,
	})
}

func (c *WatcherApp) deletePod(obj interface{}) {
	c.logger.Debug("deletePod", zap.Any("obj", obj))
	p, ok := obj.(*v1.Pod)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}
	c.queue.Add(&KubernetesEvent{
		IPs:       []string{p.Status.PodIP},
		EventType: watch.Deleted,
	})
}
