package core

import (
	"context"
	"fmt"

	"github.com/jawahars16/hawk8s/internal/kubeclient"
)

const (
	CPU    string = "cpu"
	Memory string = "memory"
)

//go:generate moq -rm -out kube_mock.go . Kube
type Kube interface {
	GetNodes(ctx context.Context) ([]kubeclient.Node, error)
	GetPods(ctx context.Context, node string) ([]kubeclient.Pod, error)
	GetNamespaces(ctx context.Context) ([]string, error)
	GetNode(ctx context.Context, name string) (kubeclient.Node, error)
}

type Service struct {
	kube Kube
}

func NewService(kube Kube) *Service {
	return &Service{
		kube: kube,
	}
}

func (s *Service) GetNamespaces(ctx context.Context) ([]namespace, error) {
	namespaces, err := s.kube.GetNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	return toNamespacesModel(namespaces), nil
}

func (s *Service) GetNodes(ctx context.Context) ([]node, error) {
	nodes, err := s.kube.GetNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodeResult := make([]node, 0, len(nodes))
	for _, n := range nodes {
		nodeResult = append(nodeResult, node{
			Name: n.Name,
			Info: fmt.Sprintf("CPU: %s | Mem: %s", cpuMilliToHumanReadable(n.AvailableCPU), memoryBytesToHumanReadable(n.AllocatableMemory)),
		})
	}
	return nodeResult, nil
}

func (s *Service) GetPods(ctx context.Context, n string) ([]pod, error) {
	pods, podErr := s.kube.GetPods(ctx, n)
	if podErr != nil && pods == nil {
		return nil, podErr
	}

	node, err := s.kube.GetNode(ctx, n)
	if err != nil {
		return nil, err
	}

	podResult := make([]pod, 0, len(pods))
	for _, p := range pods {
		podResult = append(podResult, toPodModel(p, node))
	}
	return podResult, podErr
}

func toPodModel(p kubeclient.Pod, n kubeclient.Node) pod {
	podMemory := float32(p.MemoryUsage) / float32(n.AllocatableMemory) * 100
	if podMemory <= 0.5 {
		podMemory = 0.5
	}

	podCPU := float32(p.CPUUsage) / float32(n.AvailableCPU) * 100
	if podCPU <= 0.5 {
		podCPU = 0.5
	}

	return pod{
		Name:        p.Name,
		CpuSize:     fmt.Sprintf("%v%%", podCPU),
		MemorySize:  fmt.Sprintf("%v%%", podMemory),
		CpuUsage:    cpuMilliToHumanReadable(p.CPUUsage),
		MemoryUsage: memoryBytesToHumanReadable(p.MemoryUsage),
		Color:       namespaceByName(p.Namespace).Color,
		Status:      p.Status,
		Namespace:   p.Namespace,
	}
}

func toNamespacesModel(namespaces []string) []namespace {
	var namespaceList []namespace
	for _, ns := range namespaces {
		namespaceList = append(namespaceList, namespaceByName(ns))
	}
	return namespaceList
}

func memoryBytesToHumanReadable(memoryBytes int64) string {
	// Define suffixes for different unit scales
	// suffixes := []string{"Ei", "Pi", "Ti", "Gi", "Mi", "Ki", ""}
	suffixes := []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}

	// Select the appropriate suffix and convert bytes to the corresponding unit
	i := 0
	for memoryBytes >= 1024 && i < len(suffixes)-1 {
		memoryBytes /= 1024
		i++
	}

	// Format the string with two decimal places and the chosen suffix
	return fmt.Sprintf("%d%s", memoryBytes, suffixes[i])
}

func cpuMilliToHumanReadable(cpuMilli int64) string {
	if cpuMilli < 1000 {
		return fmt.Sprintf("%dm", cpuMilli)
	}

	return fmt.Sprintf("%d", cpuMilli/1000)
}
