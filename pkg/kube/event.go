package kube

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"time"
)

type EnhancedEvent struct {
	corev1.Event   `json:",inline"`
	InvolvedObject EnhancedObjectReference `json:"involvedObject"`
}


// TODO(vsbus): restore this part but for BQ sink find a way to convert map into list of KV pairs before marshaling to JSON.
type EnhancedObjectReference struct {
	corev1.ObjectReference `json:",inline"`
	Labels      map[string]string `json:"-"`
	Annotations map[string]string `json:"-"`
}

// ToJSON does not return an error because we are %99 confident it is JSON serializable.
// TODO(makin) Is it a bad practice? It's open to discussion.
func (e *EnhancedEvent) ToJSON() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e *EnhancedEvent) GetTimestampMs() int64 {
	return e.FirstTimestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
