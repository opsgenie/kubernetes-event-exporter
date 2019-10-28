package kube

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
)

type EnhancedEvent struct {
	corev1.Event
	InvolvedObject EnhancedObjectReference
}

type EnhancedObjectReference struct {
	corev1.ObjectReference
	// TODO(makin) Should we also get its annotations? But last-applied-configuration should be dropped, what else?
	Labels map[string]string
}

// ToJSON does not return an error because we are %99 confident it is JSON serializable.
// TODO(makin) Is it a bad practice? It's open to discussion.
func (e *EnhancedEvent) ToJSON() []byte {
	b, _ := json.Marshal(e)
	return b
}
