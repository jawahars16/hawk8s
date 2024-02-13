package kubeclient

import (
	"context"
	"math/rand"
)

type MockKube struct {
}

func NewMockKube() *MockKube {
	return &MockKube{}
}

func (m *MockKube) GetNodes(ctx context.Context) ([]Node, error) {
	return []Node{
		{
			Name:              "node1",
			Status:            "Ready",
			AllocatableMemory: 1000,
			AvailableCPU:      3,
			TotalMemory:       1024,
		},
		{
			Name:              "node2",
			Status:            "Ready",
			AllocatableMemory: 900,
			AvailableCPU:      4,
			TotalMemory:       1024,
		},
		{
			Name:              "node3",
			Status:            "Ready",
			AllocatableMemory: 900,
			AvailableCPU:      1,
			TotalMemory:       1024,
		},
	}, nil
}

func (m *MockKube) GetPods(ctx context.Context) ([]Pod, error) {
	pods := make(map[string][]Pod)
	pods["default"] = []Pod{
		{
			Name:        "pod1",
			Node:        "node1",
			Namespace:   "default",
			MemoryUsage: 100,
			CPUUsage:    1,
		},
		{
			Name:      "pod2",
			Node:      "node1",
			Namespace: "default",
		},
		{
			Name:      "pod3",
			Node:      "node1",
			Namespace: "default",
		},
	}
	pods["namespace-1"] = []Pod{
		{
			Name:      "namespace-1-pod1",
			Node:      "node2",
			Namespace: "namespace-1",
		},
		{
			Name:      "namespace-1-pod2",
			Node:      "node2",
			Namespace: "namespace-1",
		},
		{
			Name:      "namespace-1-pod3",
			Node:      "node1",
			Namespace: "namespace-1",
		},
		{
			Name:      "namespace-1-pod4",
			Node:      "node3",
			Namespace: "namespace-1",
		},
	}
	var podsResult []Pod
	for _, p := range pods {
		podsResult = append(podsResult, p...)
	}
	return podsResult, nil
}

func (m *MockKube) GetNamespaces(ctx context.Context) ([]string, error) {
	return []string{"default", "namespace-1"}, nil
}

func (m *MockKube) GetPodMetrics(ctx context.Context, podName, namespace string) (int64, int64, error) {
	memory := int64(rand.Intn(500))
	cpu := int64(rand.Intn(3))
	return memory, cpu, nil
}
