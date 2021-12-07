package kube

import (
	"context"

	lru "github.com/hashicorp/golang-lru"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type NamespaceLabelCache struct {
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset

	cache *lru.ARCCache
}

func NewNamespaceLabelCache(kubeconfig *rest.Config) *NamespaceLabelCache {
	cache, err := lru.NewARC(1024)
	if err != nil {
		panic("cannot init cache: " + err.Error())
	}
	return &NamespaceLabelCache{
		dynClient: dynamic.NewForConfigOrDie(kubeconfig),
		clientset: kubernetes.NewForConfigOrDie(kubeconfig),
		cache:     cache,
	}
}

func (n *NamespaceLabelCache) GetNamespaceLabelsWithCache(nsName string) (map[string]string, error) {
	if val, ok := n.cache.Get(nsName); ok {
		return val.(map[string]string), nil
	}

	ns, err := n.clientset.CoreV1().Namespaces().Get(context.Background(), nsName, metav1.GetOptions{})
	if err == nil {
		nsLabels := ns.GetLabels()
		n.cache.Add(nsName, nsLabels)
		return nsLabels, nil
	}

	if errors.IsNotFound(err) {
		var empty map[string]string
		n.cache.Add(nsName, empty)
		return nil, nil
	}

	return nil, err
}
