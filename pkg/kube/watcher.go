package kube

import (
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type EventHandler func(event *EnhancedEvent)

type EventWatcher struct {
	informer   cache.SharedInformer
	stopper    chan struct{}
	labelCache *LabelCache
	fn         EventHandler
}

func NewEventWatcher(config *rest.Config, fn EventHandler) *EventWatcher {
	clientset := kubernetes.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events().Informer()

	watcher := &EventWatcher{
		informer:   informer,
		stopper:    make(chan struct{}),
		labelCache: NewLabelCache(config),
		fn:         fn,
	}

	informer.AddEventHandler(watcher)

	return watcher
}

func (e *EventWatcher) OnAdd(obj interface{}) {
	event := obj.(*corev1.Event)
	e.onEvent(event)
}

func (e *EventWatcher) OnUpdate(oldObj, newObj interface{}) {
	event := newObj.(*corev1.Event)
	e.onEvent(event)
}

func (e *EventWatcher) onEvent(event *corev1.Event) {
	// TODO: Re-enable this after development
	// It's probably an old event we are catching, it's not the best way but anyways
	if time.Now().Sub(event.CreationTimestamp.Time) > time.Second*5 {
		return
	}

	log.Debug().
		Str("msg", event.Message).
		Str("namespace", event.Namespace).
		Str("reason", event.Reason).
		Msg("Received event")

	ev := &EnhancedEvent{
		Event: *event.DeepCopy(),
	}

	labels, err := e.labelCache.GetLabelsWithCache(&event.InvolvedObject)
	if err != nil {
		log.Error().Err(err).Msg("Cannot list labels of the object")
		// Ignoring error, but log it anyways
	} else {
		ev.InvolvedObject.Labels = labels
		ev.InvolvedObject.ObjectReference = *event.InvolvedObject.DeepCopy()
	}

	e.fn(ev)
	return
}

func (e *EventWatcher) OnDelete(obj interface{}) {
	// Ignore deletes
}

func (e *EventWatcher) Start() {
	go e.informer.Run(e.stopper)
}

func (e *EventWatcher) Stop() {
	e.stopper <- struct{}{}
	close(e.stopper)
}
