package core_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jawahars16/hawk8s/internal/core"
	"github.com/jawahars16/hawk8s/internal/kubeclient"
	"github.com/stretchr/testify/assert"
)

func Test_GetPods(t *testing.T) {
	t.Run("given kubeclient returns non empty pods and non nil error, then return pods and error", func(t *testing.T) {
		kube := &core.KubeMock{
			GetPodsFunc: func(ctx context.Context, node string) ([]kubeclient.Pod, error) {
				assert.Equal(t, "node1", node)
				return []kubeclient.Pod{{Name: "pod1"}}, fmt.Errorf("error")
			},
			GetNodeFunc: func(ctx context.Context, name string) (kubeclient.Node, error) {
				assert.Equal(t, "node1", name)
				return kubeclient.Node{}, nil
			},
		}
		service := core.NewService(kube)
		pods, err := service.GetPods(context.Background(), "node1")
		assert.NotNil(t, pods)
		assert.Equal(t, 1, len(pods))
		assert.NotNil(t, err)
	})

	t.Run("given kubeclient returns nil pods and non nil error, then return nil pods and error", func(t *testing.T) {
		kube := &core.KubeMock{
			GetPodsFunc: func(ctx context.Context, node string) ([]kubeclient.Pod, error) {
				assert.Equal(t, "node1", node)
				return nil, fmt.Errorf("error")
			},
			GetNodeFunc: func(ctx context.Context, name string) (kubeclient.Node, error) {
				assert.Equal(t, "node1", name)
				return kubeclient.Node{}, nil
			},
		}
		service := core.NewService(kube)
		pods, err := service.GetPods(context.Background(), "node1")
		assert.Nil(t, pods)
		assert.NotNil(t, err)
	})

	t.Run("given kubeclient returns non empty pods and nil error, then return pods and nil error", func(t *testing.T) {
		kube := &core.KubeMock{
			GetPodsFunc: func(ctx context.Context, node string) ([]kubeclient.Pod, error) {
				assert.Equal(t, "node1", node)
				return []kubeclient.Pod{{Name: "pod1"}}, nil
			},
			GetNodeFunc: func(ctx context.Context, name string) (kubeclient.Node, error) {
				assert.Equal(t, "node1", name)
				return kubeclient.Node{}, nil
			},
		}
		service := core.NewService(kube)
		pods, err := service.GetPods(context.Background(), "node1")
		assert.NotNil(t, pods)
		assert.Equal(t, 1, len(pods))
		assert.Nil(t, err)
	})
}
