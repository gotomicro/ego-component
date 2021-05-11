package ek8s

import (
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *WatcherApp) addEndpoints(obj interface{}) {
	c.logger.Debug("addEndpoints", zap.Any("obj", obj))
	p, ok := obj.(*v1.Endpoints)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}

	addresses := make([]string, 0)
	for _, subsets := range p.Subsets {
		for _, address := range subsets.Addresses {
			addresses = append(addresses, address.IP)
		}
	}

	c.queue.Add(&KubernetesEvent{
		EventType: watch.Added,
		IPs:       addresses,
	})
}

func (c *WatcherApp) updateEndpoints(oldObj, newObj interface{}) {
	c.logger.Debug("updateEndpoints", zap.Any("oldObj", oldObj), zap.Any("newObj", newObj))

	op, ok := oldObj.(*v1.Endpoints)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", oldObj)
		return
	}
	np, ok := newObj.(*v1.Endpoints)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", newObj)
		return
	}
	if op.GetResourceVersion() == np.GetResourceVersion() {
		return
	}

	addresses := make([]string, 0)
	for _, subsets := range np.Subsets {
		for _, address := range subsets.Addresses {
			addresses = append(addresses, address.IP)
		}
	}

	c.queue.Add(&KubernetesEvent{
		IPs:       addresses,
		EventType: watch.Modified,
	})
}

func (c *WatcherApp) deleteEndpoints(obj interface{}) {
	c.logger.Debug("deleteEndpoints", zap.Any("obj", obj))
	p, ok := obj.(*v1.Endpoints)
	if !ok {
		c.logger.Warnf("pod-informer got object %T not *v1.Pod", obj)
		return
	}

	addresses := make([]string, 0)
	for _, subsets := range p.Subsets {
		for _, address := range subsets.Addresses {
			addresses = append(addresses, address.IP)
		}
	}

	c.queue.Add(&KubernetesEvent{
		IPs:       addresses,
		EventType: watch.Deleted,
	})
}
