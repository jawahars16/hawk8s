package kubeclient

import (
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type store struct {
	namespaces       []string
	nodes            []Node
	pods             []Pod
	podsLastModified int64
	errors           map[string]error
	lock             sync.RWMutex
}

func NewStore() *store {
	return &store{
		namespaces: make([]string, 0),
		nodes:      make([]Node, 0),
		pods:       make([]Pod, 0),
		errors:     make(map[string]error),
		lock:       sync.RWMutex{},
	}
}

func (s *store) SetError(key string, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

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
	s.lock.RLock()
	defer s.lock.RUnlock()

	if err, found := s.errors["ns"]; found {
		return nil, err
	}
	return s.namespaces, nil
}

func (s *store) AddNode(n *corev1.Node) {
	var status string
	if len(n.Status.Conditions) > 0 {
		status = string(n.Status.Conditions[0].Type)
	}
	s.nodes = append(s.nodes, Node{
		Name:              n.Name,
		Status:            status,
		AllocatableMemory: n.Status.Allocatable.Memory().Value(),
		TotalMemory:       n.Status.Capacity.Memory().Value(),
		AvailableCPU:      n.Status.Allocatable.Cpu().MilliValue(),
	})
}

func (s *store) GetNodes() ([]Node, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

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
	s.podsLastModified = time.Now().Unix()
}

func (s *store) ModifyPod(p *corev1.Pod) {
	for i, pod := range s.pods {
		if pod.Name == p.Name {
			s.pods[i].Node = p.Spec.NodeName
			s.pods[i].Status = string(p.Status.Phase)
			break
		}
	}
	s.podsLastModified = time.Now().Unix()
}

func (s *store) GetPods(node string) ([]Pod, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	err := s.errors["pods"]
	if node == "" {
		return s.pods, err
	}

	var result []Pod
	for _, pod := range s.pods {
		if pod.Node == node {
			result = append(result, pod)
		}
	}
	return result, err
}

func (s *store) DeletePod(name string) {
	for i, pod := range s.pods {
		if pod.Name == name {
			s.pods = append(s.pods[:i], s.pods[i+1:]...)
			break
		}
	}
	s.podsLastModified = time.Now().Unix()
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

			s.pods[i].CPUUsage = cpu.MilliValue()
			s.pods[i].MemoryUsage = memory.Value()
		}
	}
}
