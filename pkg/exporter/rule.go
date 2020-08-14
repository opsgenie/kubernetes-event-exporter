package exporter

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"regexp"
)

// matchString is a method to clean the code. Error handling is omitted here because these
// rules are validated before use. According to regexp.MatchString, the only way it fails its
// that the pattern does not compile.
func matchString(pattern, s string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// Rule is for matching an event
type Rule struct {
	Labels      map[string]string
	Annotations map[string]string
	Message     string
	APIVersion  string
	Kind        string
	Namespace   string
	Reason      string
	Type        string
	MinCount    int32
	Component   string
	Host        string
	Receiver    string
}

// MatchesEvent compares the rule to an event and returns a boolean value to indicate
// whether the event is compatible with the rule. All fields are compared as regular expressions
// so the user must keep that in mind while writing rules.
func (r *Rule) MatchesEvent(ev *kube.EnhancedEvent) bool {
	// These rules are just basic comparison rules, if one of them fails, it means the event does not match the rule
	rules := [][2]string{
		{r.Message, ev.Message},
		{r.APIVersion, ev.InvolvedObject.APIVersion},
		{r.Kind, ev.InvolvedObject.Kind},
		{r.Namespace, ev.Namespace},
		{r.Reason, ev.Reason},
		{r.Type, ev.Type},
		{r.Component, ev.Source.Component},
		{r.Host, ev.Source.Host},
	}

	for _, v := range rules {
		rule := v[0]
		value := v[1]
		if rule != "" {
			matches := matchString(rule, value)
			if !matches {
				return false
			}
		}
	}

	// Labels are also mutually exclusive, they all need to be present
	if r.Labels != nil && len(r.Labels) > 0 {
		for k, v := range r.Labels {
			if val, ok := ev.InvolvedObject.Labels[k]; !ok {
				return false
			} else {
				matches := matchString(v, val)
				if !matches {
					return false
				}
			}
		}
	}

	// Annotations are also mutually exclusive, they all need to be present
	if r.Annotations != nil && len(r.Annotations) > 0 {
		for k, v := range r.Annotations {
			if val, ok := ev.InvolvedObject.Annotations[k]; !ok {
				return false
			} else {
				matches := matchString(v, val)
				if !matches {
					return false
				}
			}
		}
	}

	// If minCount is not given via a config, it's already 0 and the count is already 1 and this passes.
	if ev.Count >= r.MinCount {
		return true
	} else {
		return false
	}

	// If it failed every step, it must match because our matchers are limiting
	return true
}
