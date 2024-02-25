package kubeclient

import (
	"context"
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type KubeClient struct {
	clientset *kubernetes.Clientset
	metrics   *metricsv.Clientset
	worker    *worker
	store     *store
}

func NewKubeClient() *KubeClient {
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

func (k *KubeClient) GetNode(ctx context.Context, name string) (Node, error) {
	return k.store.GetNode(name)
}
