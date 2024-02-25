package kubeclient

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type worker struct {
	client  *kubernetes.Clientset
	metrics *metricsv.Clientset
	store   *store
}

func NewWorker(client *kubernetes.Clientset, metrics *metricsv.Clientset, store *store) *worker {
	return &worker{
		client:  client,
		metrics: metrics,
		store:   store,
	}
}

func (w *worker) Run(ctx context.Context) {
	go w.watchNamespaces(ctx)
	go w.watchNodes(ctx)
	go w.watchPods(ctx)
	go w.watchPodMetrics(ctx)
}

func (w *worker) watchNamespaces(ctx context.Context) {
	watch, err := w.client.CoreV1().Namespaces().Watch(ctx, v1.ListOptions{})
	if err != nil {
		w.store.SetError("ns", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watch.ResultChan():
			ns, ok := event.Object.(*corev1.Namespace)
			if !ok {
				continue
			}
			if event.Type == "ADDED" {
				w.store.AddNamespace(ns.Name)
			} else if event.Type == "DELETED" {
				w.store.DeleteNamespace(ns.Name)
			}
		}
	}
}

func (w *worker) watchNodes(ctx context.Context) {
	watch, err := w.client.CoreV1().Nodes().Watch(ctx, v1.ListOptions{})
	if err != nil {
		w.store.SetError("nodes", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watch.ResultChan():
			node, ok := event.Object.(*corev1.Node)
			if !ok {
				continue
			}
			if event.Type == "ADDED" {
				w.store.AddNode(node)
			} else if event.Type == "DELETED" {
				w.store.DeleteNode(node.Name)
			}
		}
	}
}

func (w *worker) watchPods(ctx context.Context) {
	watch, err := w.client.CoreV1().Pods("").Watch(ctx, v1.ListOptions{})
	if err != nil {
		w.store.SetError("pods", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watch.ResultChan():
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}
			if event.Type == "ADDED" {
				w.store.AddPod(pod)
			} else if event.Type == "DELETED" {
				w.store.DeletePod(pod.Name)
			} else if event.Type == "MODIFIED" {
				w.store.ModifyPod(pod)
			}
		}
	}
}

func (w *worker) watchPodMetrics(ctx context.Context) {
	for {
		podMetrics, err := w.metrics.MetricsV1beta1().PodMetricses("").List(ctx, v1.ListOptions{})
		if err != nil {
			w.store.SetError("podMetrics", err)
			return
		}

		w.store.UpdateMetrics(podMetrics.Items)
		time.Sleep(5 * time.Second)
	}
}
