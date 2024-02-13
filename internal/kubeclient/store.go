package kubeclient

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type store struct {
	namespaces []string
	nodes      []Node
	pods       []Pod
	errors     map[string]error
}

func NewStore() *store {
	return &store{
		namespaces: make([]string, 0),
		nodes:      make([]Node, 0),
		pods:       make([]Pod, 0),
		errors:     make(map[string]error),
	}
}

func (s *store) SetError(key string, err error) {
	s.errors[key] = err
}

func (s *store) AddNamespace(namespace string) {
	s.namespaces = append(s.namespaces, namespace)
}

func (s *store) DeleteNamespace(namespace string) {
	for i, ns := range s.namespaces {
		if ns == namespace {
			s.namespaces = append(s.namespaces[:i], s.namespaces[i+1:]...)
			break
		}
	}
}

func (s *store) GetNamespaces() ([]string, error) {
	if err, found := s.errors["ns"]; found {
		return nil, err
	}
	return s.namespaces, nil
}

func (s *store) AddNode(n *corev1.Node) {
	s.nodes = append(s.nodes, Node{
		Name:              n.Name,
		Status:            string(n.Status.Conditions[0].Type),
		AllocatableMemory: n.Status.Allocatable.Memory().Value(),
		TotalMemory:       n.Status.Capacity.Memory().Value(),
		AvailableCPU:      n.Status.Allocatable.Cpu().ScaledValue(resource.Micro),
	})
}

func (s *store) GetNodes() ([]Node, error) {
	if err, found := s.errors["nodes"]; found {
		return nil, err
	}
	return s.nodes, nil
}

func (s *store) GetNode(name string) (Node, error) {
	for _, node := range s.nodes {
		if node.Name == name {
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("node %s not found", name)
}

func (s *store) DeleteNode(name string) {
	for i, node := range s.nodes {
		if node.Name == name {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			break
		}
	}
}

func (s *store) AddPod(p *corev1.Pod) {
	s.pods = append(s.pods, Pod{
		Name:      p.Name,
		Node:      p.Spec.NodeName,
		Namespace: p.Namespace,
		Status:    string(p.Status.Phase),
	})
}

func (s *store) ModifyPod(p *corev1.Pod) {
	for i, pod := range s.pods {
		if pod.Name == p.Name {
			s.pods[i].Node = p.Spec.NodeName
			s.pods[i].Status = string(p.Status.Phase)
			break
		}
	}
}

func (s *store) GetPods(node string) ([]Pod, error) {
	if err, found := s.errors["pods"]; found {
		return nil, err
	}

	if node == "" {
		return s.pods, nil
	}

	var result []Pod
	for _, pod := range s.pods {
		if pod.Node == node {
			result = append(result, pod)
		}
	}
	// if err, found := s.errors["podMetrics"]; found {
	// 	return result, err
	// }
	return result, nil
}

func (s *store) DeletePod(name string) {
	for i, pod := range s.pods {
		if pod.Name == name {
			s.pods = append(s.pods[:i], s.pods[i+1:]...)
			break
		}
	}
}

func (s *store) UpdateMetrics(podMetrics []v1beta1.PodMetrics) {
	metricsMap := make(map[string]v1beta1.PodMetrics)
	for _, podMetrics := range podMetrics {
		metricsMap[podMetrics.Name] = podMetrics
	}

	for i, pod := range s.pods {
		metrics, ok := metricsMap[pod.Name]
		if ok {
			cpu := metrics.Containers[0].Usage.Cpu()
			memory := metrics.Containers[0].Usage.Memory()

			s.pods[i].CPUUsage = cpu.ScaledValue(resource.Micro)
			s.pods[i].MemoryUsage = memory.Value()
		}
	}
}
