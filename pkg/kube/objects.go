package kube

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"strings"
)

func GetObject(reference *v1.ObjectReference, clientset *kubernetes.Clientset, dynClient dynamic.Interface) (*unstructured.Unstructured, error) {
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

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, err
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := rm.RESTMapping(gk, version)
	if err != nil {
		return nil, err
	}

	item, err := dynClient.
		Resource(mapping.Resource).
		Namespace(reference.Namespace).
		Get(context.Background(), reference.Name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return item, nil
}
