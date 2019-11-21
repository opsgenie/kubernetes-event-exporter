package kube

import (
	lru "github.com/hashicorp/golang-lru"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"strings"
	"sync"
)

type LabelCache struct {
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset

	// TODO: Obviously need something better with concurrency and LRU because it will grow indefinitely
	// cache map[types.UID]map[string]string

	cache *lru.ARCCache
	sync.RWMutex
}

func NewLabelCache(kubeconfig *rest.Config) (*LabelCache) {
	cache, err := lru.NewARC(1024)
	if err != nil {
		panic("cannot init cache: " + err.Error())
	}
	return &LabelCache{
		dynClient: dynamic.NewForConfigOrDie(kubeconfig),
		clientset: kubernetes.NewForConfigOrDie(kubeconfig),
		cache:     cache,
	}
}

func (l *LabelCache) GetObject(reference *v1.ObjectReference) (*unstructured.Unstructured, error) {
	var group, version string
	s := strings.Split(reference.APIVersion, "/")
	if len(s) == 1 {
		group = ""
		version = s[0]
	} else {
		group = s[0]
		version = s[1]
	}

	gk := schema.GroupKind{Group: group, Kind: reference.Kind}

	groupResources, err := restmapper.GetAPIGroupResources(l.clientset.Discovery())
	if err != nil {
		return nil, err
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := rm.RESTMapping(gk, version)
	if err != nil {
		return nil, err
	}

	item, err := l.dynClient.
		Resource(mapping.Resource).
		Namespace(reference.Namespace).
		Get(reference.Name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (l *LabelCache) GetLabelsWithCache(reference *v1.ObjectReference) (map[string]string, error) {
	uid := reference.UID

	if val, ok := l.cache.Get(uid); ok {
		return val.(map[string]string), nil
	}

	obj, err := l.GetObject(reference)
	if err == nil {
		labels := obj.GetLabels()
		l.cache.Add(uid, labels)
		return labels, nil
	}

	if errors.IsNotFound(err) {
		// There can be events without the involved objects existing, they seem to be not garbage collected?
		// Marking it nil so that we can return faster
		var empty map[string]string
		l.cache.Add(uid, empty)
		return nil, nil
	}

	// An non-ignorable error occurred
	return nil, err
}
