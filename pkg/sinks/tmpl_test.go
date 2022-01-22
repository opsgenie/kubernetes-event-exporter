package sinks

import (
	"testing"
	"time"

	"github.com/clbanning/mxj"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestEvent() *kube.EnhancedEvent {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	ev.Type = "Warning"
	ev.Count = 23
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Name = "nginx-server-123abc-456def"
	ev.Message = "Successfully pulled image \"nginx:latest\""
	ev.FirstTimestamp = v1.Time{Time: time.Now()}
	return ev
}

func TestLayoutConvert(t *testing.T) {
	ev := newTestEvent()

	// Because Go, when parsing yaml, its []interface, not []string
	var tagz interface{}
	tagz = make([]interface{}, 2)
	tagz.([]interface{})[0] = "sre"
	tagz.([]interface{})[1] = "ops"

	layout := map[string]interface{}{
		"details": map[interface{}]interface{}{
			"message":   "{{ .Message }}",
			"kind":      "{{ .InvolvedObject.Kind }}",
			"name":      "{{ .InvolvedObject.Name }}",
			"namespace": "{{ .Namespace }}",
			"type":      "{{ .Type }}",
			"tags":      tagz,
		},
		"eventType": "kube-event",
		"region":    "us-west-2",
		"createdAt": "{{ .GetTimestampMs }}", // TODO: Test Int casts
	}

	res, err := convertLayoutTemplate(layout, ev)
	require.NoError(t, err)
	require.Equal(t, res["eventType"], "kube-event")

	val, ok := res["details"].(map[string]interface{})

	require.True(t, ok, "cannot cast to event")

	val2, ok2 := val["message"].(string)
	require.True(t, ok2, "cannot cast message to string")

	require.Equal(t, val2, ev.Message)
}

var xmltests = []struct {
	name         string
	layout       map[string]interface{}
	expectedRoot string
}{
	{
		"XML with root node",
		map[string]interface{}{
			"event": map[interface{}]interface{}{
				"title":     "{{ .InvolvedObject.Name }}",
				"severity":  "{{ .Type }}",
				"Count":     "{{ .Count }}",
				"eventType": "kube-event",
				"region":    "us-west-2",
				"createdAt": "{{ .GetTimestampMs }}",
			}},
		"event",
	},
	{
		"XML without root node",
		map[string]interface{}{
			"title":     "{{ .InvolvedObject.Name }}",
			"severity":  "{{ .Type }}",
			"Count":     "{{ .Count }}",
			"eventType": "kube-event",
			"region":    "us-west-2",
			"createdAt": "{{ .GetTimestampMs }}",
		},
		"doc",
	},
	{
		"XML format k8s event (no layout)",
		nil,
		"doc",
	},
}

func TestXMLEventWithLayout(t *testing.T) {
	ev := newTestEvent()

	for _, tt := range xmltests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := serializeXMLEventWithLayout(tt.layout, ev)

			require.NoError(t, err)

			//t.Logf("res: %v", string(res))

			m, err := mxj.NewMapXml(res, true)

			require.NoError(t, err)

			root, ok := m[tt.expectedRoot].(map[string]interface{})

			require.True(t, ok)

			if tt.layout != nil {
				require.Equal(t, root["title"], "nginx-server-123abc-456def")
				require.Equal(t, root["severity"], "Warning")
				require.Equal(t, root["region"], "us-west-2")
				require.Equal(t, root["Count"], 23.0) // mxj unmarshal try to cast only to boolean or float64
			} else {
				require.Equal(t, root["involvedObject"].(map[string]interface{})["name"], "nginx-server-123abc-456def")
				require.Equal(t, root["type"], "Warning")
				require.Equal(t, root["count"], 23.0) // mxj unmarshal try to cast only to boolean or float64
			}

		})
	}
}
