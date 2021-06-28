package kube

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestEnhancedEvent_DeDot(t *testing.T) {
	tests := []struct {
		name string
		in   EnhancedEvent
		want EnhancedEvent
	}{
		{
			name: "nothing",
			in: EnhancedEvent{
				Event: corev1.Event{
					Message: "foovar",
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"test": "bar"},
					},
				},
				InvolvedObject: EnhancedObjectReference{
					Labels: map[string]string{"faz": "var"},
				},
			},
			want: EnhancedEvent{
				Event: corev1.Event{
					Message: "foovar",
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"test": "bar"},
					},
				},
				InvolvedObject: EnhancedObjectReference{
					Labels: map[string]string{"faz": "var"},
				},
			},
		},
		{
			name: "dedot",
			in: EnhancedEvent{
				Event: corev1.Event{
					Message: "foovar",
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"test.io.": "bar"},
					},
				},
				InvolvedObject: EnhancedObjectReference{
					Labels: map[string]string{"faz.net": "var"},
				},
			},
			want: EnhancedEvent{
				Event: corev1.Event{
					Message: "foovar",
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"test_io_": "bar"},
					},
				},
				InvolvedObject: EnhancedObjectReference{
					Labels: map[string]string{"faz_net": "var"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.DeDot()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestEnhancedEvent_DeDot_MustNotAlternateOriginal(t *testing.T) {
	expected := EnhancedEvent{
		Event: corev1.Event{
			Message: "foovar",
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{"test.io": "bar"},
				Labels:      map[string]string{"faz.net": "var"},
			},
		},
		InvolvedObject: EnhancedObjectReference{
			Annotations: map[string]string{"test.io": "bar"},
			Labels:      map[string]string{"faz.net": "var"},
		},
	}
	in := EnhancedEvent{
		Event: corev1.Event{
			Message: "foovar",
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{"test.io": "bar"},
				Labels:      map[string]string{"faz.net": "var"},
			},
		},
		InvolvedObject: EnhancedObjectReference{
			Annotations: map[string]string{"test.io": "bar"},
			Labels:      map[string]string{"faz.net": "var"},
		},
	}
	in.DeDot()
	assert.EqualValues(t, expected, in)
}
