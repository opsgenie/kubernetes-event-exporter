package exporter

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/sinks"
	"github.com/stretchr/testify/assert"
	"testing"
)

// testReceiverRegistry just records the events to the registry so that tests can validate routing behavior
type testReceiverRegistry struct {
	rcvd map[string][]*kube.EnhancedEvent
}

func (t *testReceiverRegistry) Register(string, sinks.Sink) {
	panic("Why do you call this? It's for counting imaginary events for tests only")
}

func (t *testReceiverRegistry) SendEvent(name string, event *kube.EnhancedEvent) {
	if t.rcvd == nil {
		t.rcvd = make(map[string][]*kube.EnhancedEvent)
	}

	if _, ok := t.rcvd[name]; !ok {
		t.rcvd[name] = make([]*kube.EnhancedEvent, 0)
	}

	t.rcvd[name] = append(t.rcvd[name], event)
}

func (t *testReceiverRegistry) Close() {
	// No-op
}

func (t *testReceiverRegistry) isEventRcvd(name string, event *kube.EnhancedEvent) bool {
	if val, ok := t.rcvd[name]; !ok {
		return false
	} else {
		for _, v := range val {
			if v == event {
				return true
			}
		}
		return false
	}
}

func (t *testReceiverRegistry) count(name string) int {
	if val, ok := t.rcvd[name]; ok {
		return len(val)
	} else {
		return 0
	}
}

func TestEmptyRoute(t *testing.T) {
	ev := kube.EnhancedEvent{}
	reg := testReceiverRegistry{}

	r := Route{}

	r.ProcessEvent(&ev, &reg)
	assert.Empty(t, reg.rcvd)
}

func TestBasicRoute(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Match: []Rule{{
			Namespace: "kube-system",
			Receiver:  "osman",
		}},
	}

	r.ProcessEvent(&ev, &reg)
	assert.True(t, reg.isEventRcvd("osman", &ev))
}

func TestDropRule(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Drop: []Rule{{
			Namespace: "kube-system",
		}},
		Match: []Rule{{
			Receiver: "osman",
		}},
	}

	r.ProcessEvent(&ev, &reg)
	assert.False(t, reg.isEventRcvd("osman", &ev))
	assert.Zero(t, reg.count("osman"))
}

func TestSingleLevelMultipleMatchRoute(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Match: []Rule{{
			Namespace: "kube-system",
			Receiver:  "osman",
		}, {
			Receiver: "any",
		}},
	}

	r.ProcessEvent(&ev, &reg)
	assert.True(t, reg.isEventRcvd("osman", &ev))
	assert.True(t, reg.isEventRcvd("any", &ev))
}

func TestSubRoute(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Match: []Rule{{
			Namespace: "kube-system",
		}},
		Routes: []Route{{
			Match: []Rule{{
				Receiver: "osman",
			}},
		}},
	}

	r.ProcessEvent(&ev, &reg)

	assert.True(t, reg.isEventRcvd("osman", &ev))
}

func TestSubSubRoute(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Match: []Rule{{
			Namespace: "kube-*",
		}},
		Routes: []Route{{
			Match: []Rule{{
				Receiver: "osman",
			}},
			Routes: []Route{{
				Match: []Rule{{
					Receiver: "any",
				}},
			}},
		}},
	}

	r.ProcessEvent(&ev, &reg)

	assert.True(t, reg.isEventRcvd("osman", &ev))
	assert.True(t, reg.isEventRcvd("any", &ev))
}

func TestSubSubRouteWithDrop(t *testing.T) {
	ev := kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	reg := testReceiverRegistry{}

	r := Route{
		Match: []Rule{{
			Namespace: "kube-*",
		}},
		Routes: []Route{{
			Match: []Rule{{
				Receiver: "osman",
			}},
			Routes: []Route{{
				Drop: []Rule{{
					Namespace: "kube-system",
				}},
				Match: []Rule{{
					Receiver: "any",
				}},
			}},
		}},
	}

	r.ProcessEvent(&ev, &reg)

	assert.True(t, reg.isEventRcvd("osman", &ev))
	assert.False(t, reg.isEventRcvd("any", &ev))
}

// Test for issue: https://github.com/opsgenie/kubernetes-event-exporter/issues/51
func Test_GHIssue51(t *testing.T) {
	ev1 := kube.EnhancedEvent{}
	ev1.Type = "Warning"
	ev1.Reason = "FailedCreatePodContainer"

	ev2 := kube.EnhancedEvent{}
	ev2.Type = "Warning"
	ev2.Reason = "FailedCreate"

	reg := testReceiverRegistry{}

	r := Route{
		Drop: []Rule{{
			Type: "Normal",
		}},
		Match: []Rule{{
			Reason: "FailedCreatePodContainer",
			Receiver: "elastic",
		}},
	}

	r.ProcessEvent(&ev1, &reg)
	r.ProcessEvent(&ev2, &reg)

	assert.True(t, reg.isEventRcvd("elastic", &ev1))
	assert.False(t, reg.isEventRcvd("elastic", &ev2))
}