package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	if err != nil {
		log.Fatal(err)
	}

	clientset := kubernetes.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	// Kubernetes serves an utility to handle API crashes
	defer runtime.HandleCrash()

	// This is the part where your custom code gets triggered based on the
	// event that the shared informer catches

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// When a new pod gets created
		AddFunc: func(obj interface{}) {
			event := obj.(*corev1.Event)
			// It's probably an old event we are catching
			if time.Now().Sub(event.CreationTimestamp.Time) > time.Second*10 {
				return
			}
			fmt.Println("add", event.Namespace, event.Reason, event.Message)
		},
		// When a pod gets updated
		UpdateFunc: func(oldObj interface{}, obj interface{}) {
			event := obj.(*corev1.Event)
			fmt.Println("update",  event.Namespace, event.Reason, event.Message, event.Count)
		},
		// When a pod gets deleted
		DeleteFunc: func(interface{}) {

		},
	})

	// You need to start the informer, in my case, it runs in the background
	go informer.Run(stopper)

	time.Sleep(time.Minute * 10)
}
