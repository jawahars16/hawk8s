package kubeclient

import (
	"context"
	"flag"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type KubeClient struct {
	clientset *kubernetes.Clientset
	metrics   *metricsv.Clientset
	cache     cache
	worker    *worker
	store     *store
}

func NewKubeClient(cache cache) *KubeClient {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	store := NewStore()
	worker := NewWorker(clientset, metricsClientset, store)
	worker.Run(context.Background())

	return &KubeClient{
		clientset: clientset,
		metrics:   metricsClientset,
		cache:     cache,
		worker:    worker,
		store:     store,
	}
}

func (k *KubeClient) GetNodes(ctx context.Context) ([]Node, error) {
	return k.store.GetNodes()
}

func (k *KubeClient) GetNamespaces(ctx context.Context) ([]string, error) {
	return k.store.GetNamespaces()
}

func (k *KubeClient) GetPods(ctx context.Context, node string) ([]Pod, error) {
	return k.store.GetPods(node)
}

func (k *KubeClient) GetPods1(ctx context.Context) ([]Pod, error) {
	var errFinal error
	var podResult []Pod = make([]Pod, 0)
	var podList *corev1.PodList

	cachedPods, found := k.cache.Get("pods")
	if found {
		podList = cachedPods.(*corev1.PodList)
	} else {
		var err error
		pods := k.clientset.CoreV1().Pods("")
		podList, err = pods.List(ctx, v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		k.cache.Set("pods", podList)
	}

	podMetricsList, err := k.metrics.MetricsV1beta1().PodMetricses("").List(ctx, v1.ListOptions{})
	if err != nil {
		errFinal = err
	}

	if err == nil {
		podMetricsMap := make(map[string]v1beta1.PodMetrics)
		for _, podMetrics := range podMetricsList.Items {
			podMetricsMap[podMetrics.Name] = podMetrics
		}

		for _, pod := range podList.Items {
			var memeoryUsage int64
			var cpuUsage int64
			podMetrics, ok := podMetricsMap[pod.Name]
			if ok && podMetrics.Containers != nil && len(podMetrics.Containers) > 0 {
				cpu := podMetrics.Containers[0].Usage.Cpu()
				memory := podMetrics.Containers[0].Usage.Memory()

				cpuUsage = cpu.ScaledValue(resource.Micro)
				memeoryUsage = memory.Value()
			}

			podResult = append(podResult, Pod{
				Name:        pod.Name,
				Node:        pod.Spec.NodeName,
				Namespace:   pod.Namespace,
				MemoryUsage: memeoryUsage,
				CPUUsage:    cpuUsage,
			})
		}
		return podResult, errFinal
	}

	for _, pod := range podList.Items {
		podResult = append(podResult, Pod{
			Name:      pod.Name,
			Node:      pod.Spec.NodeName,
			Namespace: pod.Namespace,
		})
	}
	return podResult, errFinal
}

func (k *KubeClient) GetNode(ctx context.Context, name string) (Node, error) {
	return k.store.GetNode(name)
}
