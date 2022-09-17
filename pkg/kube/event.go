package kube

import (
	"encoding/json"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type EnhancedEvent struct {
	corev1.Event   `json:",inline"`
	InvolvedObject EnhancedObjectReference `json:"involvedObject"`
}

// DeDot replaces all dots in the labels and annotations with underscores. This is required for example in the
// elasticsearch sink. The dynamic mapping generation interprets dots in JSON keys as as path in a onject.
// For reference see this logstash filter: https://www.elastic.co/guide/en/logstash/current/plugins-filters-de_dot.html
func (e EnhancedEvent) DeDot() EnhancedEvent {
	c := e
	c.Labels = dedotMap(e.Labels)
	c.Annotations = dedotMap(e.Annotations)
	c.InvolvedObject.Labels = dedotMap(e.InvolvedObject.Labels)
	c.InvolvedObject.Annotations = dedotMap(e.InvolvedObject.Annotations)
	return c
}

func dedotMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return in
	}
	ret := make(map[string]string, len(in))
	for key, value := range in {
		nKey := strings.ReplaceAll(key, ".", "_")
		ret[nKey] = value
	}
	return ret
}

type EnhancedObjectReference struct {
	corev1.ObjectReference `json:",inline"`
	Labels                 map[string]string `json:"labels,omitempty"`
	Annotations            map[string]string `json:"annotations,omitempty"`
}

// ToJSON does not return an error because we are %99 confident it is JSON serializable.
// TODO(makin) Is it a bad practice? It's open to discussion.
func (e *EnhancedEvent) ToJSON() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e *EnhancedEvent) GetTimestampMs() int64 {
	timestamp := e.FirstTimestamp.Time
	if timestamp.IsZero() {
		timestamp = e.EventTime.Time
	}

	return timestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (e *EnhancedEvent) GetTimestampISO8601() string {
	timestamp := e.FirstTimestamp.Time
	if timestamp.IsZero() {
		timestamp = e.EventTime.Time
	}

	layout := "2006-01-02T15:04:05.000Z"
	return timestamp.Format(layout)
}
