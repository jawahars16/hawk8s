package core

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/jawahars16/kubemon/internal/kubeclient"
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
			Info: fmt.Sprintf("CPU: %s | Mem: %s", toReadableCPU(n.AvailableCPU), toReadableMemory(n.AllocatableMemory)),
		})
	}
	return nodeResult, nil
}

func (s *Service) GetPods(ctx context.Context, n string) ([]pod, error) {
	pods, err := s.kube.GetPods(ctx, n)
	if err != nil {
		return nil, err
	}

	node, err := s.kube.GetNode(ctx, n)
	if err != nil {
		return nil, err
	}

	podResult := make([]pod, 0, len(pods))
	for _, p := range pods {
		podResult = append(podResult, toPodModel(p, node))
	}
	return podResult, nil
}

// func (n *Service) GetViewModel(ctx context.Context, namespace string, md string) *viewModel {
// 	vm := &viewModel{}

// 	namespaces, err := n.kube.GetNamespaces(ctx)
// 	if err != nil {
// 		vm.Error = err.Error()
// 		return vm
// 	}
// 	vm.Namespaces = toNamespacesModel(namespaces)

// 	nodes, err := n.kube.GetNodes(ctx)
// 	if err != nil {
// 		vm.Error = err.Error()
// 		return vm
// 	}

// 	pods, err := n.kube.GetPods(ctx)
// 	if err != nil {
// 		vm.Error = err.Error()
// 	}

// 	if pods == nil {
// 		return vm
// 	}

// 	var title string
// 	if md == CPU {
// 		title = "Showing CPU usage"
// 	} else {
// 		title = "Showing Memory usage"
// 	}
// 	viewModel := viewModel{
// 		ActiveNamespace: namespace,
// 		ActiveMode:      md,
// 		Nodes:           mapToViewModel(nodes, pods, namespace, md),
// 		Title:           fmt.Sprintf("%s for %s namespace", title, namespace),
// 		Namespaces:      vm.Namespaces,
// 		Error:           vm.Error,
// 		Modes: []mode{
// 			{Name: "CPU", Value: CPU},
// 			{Name: "Memory", Value: Memory},
// 		},
// 	}
// 	return &viewModel
// }

func mapToViewModel(nodes []kubeclient.Node, pods []kubeclient.Pod, ns string, mode string) []node {
	var nodeList []node
	for _, n := range nodes {
		node := toNodeModel(n, pods, ns, mode)
		nodeList = append(nodeList, node)
	}
	return nodeList
}

func toNodeModel(n kubeclient.Node, pods []kubeclient.Pod, ns string, mode string) node {
	node := node{
		Name: n.Name,
	}
	for _, p := range pods {
		if p.Node == n.Name {
			pod := toPodModel(p, n)
			node.Pods = append(node.Pods, pod)
		}
	}

	var info string
	if mode == "cpu" {
		info = fmt.Sprintf("CPU: %s | %d pods", toReadableCPU(n.AvailableCPU), len(node.Pods))
	} else {
		info = fmt.Sprintf("Mem: %s | %d pods", toReadableMemory(n.AllocatableMemory), len(node.Pods))
	}

	node.Info = info
	return node
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
		CpuUsage:    toReadableCPU(p.CPUUsage),
		MemoryUsage: toReadableMemory(p.MemoryUsage),
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

func toReadableMemory(size int64) string {
	var (
		suffixes [5]string
	)
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"

	if size <= 0 {
		return "0 B"
	}

	base := math.Log(float64(size)) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + string(getSuffix)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func toReadableCPU(size int64) string {
	cpu := float64(size) / 1_000_000
	if cpu < 1 {
		cpu = float64(size) / 1_000
		return fmt.Sprintf("%.2fm", cpu)
	}
	return fmt.Sprintf("%.2f", cpu)
}
