package kubeclient_test

import (
	"fmt"
	"testing"

	"github.com/jawahars16/hawk8s/internal/kubeclient"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_Store(t *testing.T) {
	t.Run("Set error", func(t *testing.T) {
		store := kubeclient.NewStore()
		store.SetError("ns", fmt.Errorf("error"))
		nss, err := store.GetNamespaces()
		assert.Nil(t, nss)
		assert.NotNil(t, err)
	})

	t.Run("Add namespace", func(t *testing.T) {
		store := kubeclient.NewStore()
		store.AddNamespace("ns1")
		nss, err := store.GetNamespaces()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(nss))
		assert.Equal(t, "ns1", nss[0])
	})

	t.Run("Delete namespace", func(t *testing.T) {
		store := kubeclient.NewStore()
		store.AddNamespace("ns1")
		store.DeleteNamespace("ns1")
		nss, err := store.GetNamespaces()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(nss))
	})

	t.Run("Add node", func(t *testing.T) {
		store := kubeclient.NewStore()
		store.AddNode(&corev1.Node{})
		nodes, err := store.GetNodes()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(nodes))
	})
}
