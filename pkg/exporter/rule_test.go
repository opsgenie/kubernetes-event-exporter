package exporter

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	var r Rule

	assert.True(t, r.MatchesEvent(ev))
}

func TestBasicRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	r := Rule{
		Namespace: "kube-system",
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestBasicNoMatchRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	r := Rule{
		Namespace: "kube-system",
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestBasicRegexRule(t *testing.T) {
	ev1 := &kube.EnhancedEvent{}
	ev1.Namespace = "kube-system"

	ev2 := &kube.EnhancedEvent{}
	ev2.Namespace = "kube-public"

	ev3 := &kube.EnhancedEvent{}
	ev3.Namespace = "default"

	r := Rule{
		Namespace: "kube-*",
	}

	assert.True(t, r.MatchesEvent(ev1))
	assert.True(t, r.MatchesEvent(ev2))
	assert.False(t, r.MatchesEvent(ev3))
}

func TestLabelRegexRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"version": "alpha-123",
	}

	r := Rule{
		Labels: map[string]string{
			"version": "alpha",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestOneLabelMatchesRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"env": "prod",
	}

	r := Rule{
		Labels: map[string]string{
			"env": "prod",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestOneLabelDoesNotMatchRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"env": "lab",
	}

	r := Rule{
		Labels: map[string]string{
			"env": "prod",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestTwoLabelMatchesRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "beta",
	}

	r := Rule{
		Labels: map[string]string{
			"env":     "prod",
			"version": "beta",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestTwoLabelRequiredRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}

	r := Rule{
		Labels: map[string]string{
			"env":     "prod",
			"version": "beta",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestTwoLabelRequiredOneMissingRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"age":     "very-old",
		"version": "beta",
	}

	r := Rule{
		Labels: map[string]string{
			"env":     "prod",
			"version": "beta",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestOneAnnotationMatchesRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Annotations = map[string]string{
		"name":    "source",
		"service": "event-exporter",
	}

	r := Rule{
		Annotations: map[string]string{
			"name": "sou*",
		},
	}
	assert.True(t, r.MatchesEvent(ev))
}

func TestOneAnnotationDoesNotMatchRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Annotations = map[string]string{
		"name": "source",
	}

	r := Rule{
		Annotations: map[string]string{
			"name": "test*",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestTwoAnnotationsMatchesRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Annotations = map[string]string{
		"name":    "source",
		"service": "event-exporter",
	}

	r := Rule{
		Annotations: map[string]string{
			"name":    "sou.*",
			"service": "event*",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestTwoAnnotationsRequiredOneMissingRule(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Annotations = map[string]string{
		"service": "event-exporter",
	}

	r := Rule{
		Annotations: map[string]string{
			"name":    "sou*",
			"service": "event*",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestComplexRuleNoMatch(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}

	r := Rule{
		Namespace: "kube-system",
		Type:      "Warning",
		Labels: map[string]string{
			"env":     "prod",
			"version": "alpha",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestComplexRuleMatches(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}
	ev.InvolvedObject.Annotations = map[string]string{
		"service": "event-exporter",
	}

	r := Rule{
		Namespace: "kube-system",
		Kind:      "Pod",
		Labels: map[string]string{
			"env":     "prod",
			"version": "alpha",
		},
		Annotations: map[string]string{
			"service": "event*",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestComplexRuleAnnotationsNoMatch(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}
	ev.InvolvedObject.Annotations = map[string]string{
		"service": "event*",
	}

	r := Rule{
		Namespace: "kube-system",
		Kind:      "Pod",
		Labels: map[string]string{
			"env":     "prod",
			"version": "alpha",
		},
		Annotations: map[string]string{
			"name": "test*",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestComplexRuleMatchesRegexp(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}

	r := Rule{
		Namespace: "kube*",
		Kind:      "Po*",
		Labels: map[string]string{
			"env":     "prod",
			"version": "alpha|beta",
		},
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestComplexRuleNoMatchRegexp(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "kube-system"
	ev.Type = "Pod"
	ev.InvolvedObject.Labels = map[string]string{
		"env":     "prod",
		"version": "alpha",
	}

	r := Rule{
		Namespace: "kube*",
		Type:      "Deployment|ReplicaSet",
		Labels: map[string]string{
			"env":     "prod",
			"version": "alpha|beta",
		},
	}

	assert.False(t, r.MatchesEvent(ev))
}

func TestMessageRegexp(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	ev.Type = "Pod"
	ev.Message = "Successfully pulled image \"nginx:latest\""

	r := Rule{
		Type:    "Pod",
		Message: "pulled.*nginx.*",
	}

	assert.True(t, r.MatchesEvent(ev))
}

func TestCount(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	ev.Type = "Pod"
	ev.Message = "Successfully pulled image \"nginx:latest\""
	ev.Count = 5

	r := Rule{
		Type:     "Pod",
		Message:  "pulled.*nginx.*",
		MinCount: 30,
	}

	assert.False(t, r.MatchesEvent(ev))
}
