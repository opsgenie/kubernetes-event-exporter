package kube

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"time"
	"unicode"
)

type EnhancedEvent struct {
	corev1.Event   `json:",inline"`
	InvolvedObject EnhancedObjectReference `json:"involvedObject"`
}

type EnhancedObjectReference struct {
	corev1.ObjectReference `json:",inline"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Returns a map filtering out keys that have nil value assigned.
func dropNils(x map[string]interface{}) map[string]interface{} {
    y := make(map[string]interface{})
    for key, value := range x {
        if value != nil {
            if mapValue, ok := value.(map[string]interface{}); ok {
                y[key] = dropNils(mapValue)
            } else {
                y[key] = value
            }
        }
    }
    return y
}

// Returns a string representing a fixed key. BigQuery expects keys to be valid identifiers, so if they aren't we modify them.
func fixKey(key string) string {
        var fixedKey string
        if !unicode.IsLetter(rune(key[0])) {
            fixedKey = "_"
        }
        for _, ch := range key {
            if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
                fixedKey = fixedKey + string(ch)
            } else {
                fixedKey = fixedKey + "_"
            }
        }
        return fixedKey
}

// Returns a map copy with fixed keys.
func fixKeys(x map[string]interface{}) map[string]interface{} {
    y := make(map[string]interface{})
    for key, value := range x {
            if mapValue, ok := value.(map[string]interface{}); ok {
                    y[fixKey(key)] = fixKeys(mapValue)
            } else {
                    y[fixKey(key)] = value
            }
    }
    return y
}

// ToJSON does not return an error because we are %99 confident it is JSON serializable.
// TODO(makin) Is it a bad practice? It's open to discussion.
func (e *EnhancedEvent) ToJSON() []byte {
        jsonBytes, _ := json.Marshal(e)
	var mapStruct map[string]interface{}
	json.Unmarshal(jsonBytes, &mapStruct)
        jsonBytes, _ = json.Marshal(fixKeys(dropNils(mapStruct)))
        return jsonBytes
}

func (e *EnhancedEvent) GetTimestampMs() int64 {
	return e.FirstTimestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
