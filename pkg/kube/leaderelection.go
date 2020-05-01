package kube

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// LeaderElectionConfig is used to enable leader election
type LeaderElectionConfig struct {
	Enabled          bool   `yaml:"enabled"`
	LeaderElectionID string `yaml:"leaderElectionID"`
}

const (
	inClusterNamespacePath  = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	defaultLeaderElectionID = "kubernetes-event-exporter"
	defaultNamespace        = "default" // this is used for local development
	defaultLeaseDuration    = 15 * time.Second
	defaultRenewDeadline    = 10 * time.Second
	defaultRetryPeriod      = 2 * time.Second
)

// NewResourceLock creates a new config map resource lock for use in a leader
// election loop
func newResourceLock(config *rest.Config, leaderElectionID string) (resourcelock.Interface, error) {
	if leaderElectionID == "" {
		leaderElectionID = defaultLeaderElectionID
	}

	leaderElectionNamespace, err := getInClusterNamespace()
	if err != nil {
		leaderElectionNamespace = defaultNamespace
	}

	// Leader id, needs to be unique
	id, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	id = id + "_" + string(uuid.NewUUID())

	// Construct client for leader election
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return resourcelock.New(resourcelock.ConfigMapsResourceLock,
		leaderElectionNamespace,
		leaderElectionID,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
}

func getInClusterNamespace() (string, error) {
	// Check whether the namespace file exists.
	// If not, we are not running in cluster so can't guess the namespace.
	_, err := os.Stat(inClusterNamespacePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("not running in-cluster, please specify leaderElectionIDspace")
	} else if err != nil {
		return "", fmt.Errorf("error checking namespace file: %w", err)
	}

	// Load the namespace file and return its content
	namespace, err := ioutil.ReadFile(inClusterNamespacePath)
	if err != nil {
		return "", fmt.Errorf("error reading namespace file: %w", err)
	}
	return string(namespace), nil
}

// NewLeaderElector return  a leader elector object using client-go
func NewLeaderElector(leaderElectionID string, config *rest.Config, startFunc func(context.Context), stopFunc func()) (*leaderelection.LeaderElector, error) {
	resourceLock, err := newResourceLock(config, leaderElectionID)
	if err != nil {
		return &leaderelection.LeaderElector{}, err
	}

	l, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: defaultLeaseDuration,
		RenewDeadline: defaultRenewDeadline,
		RetryPeriod:   defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: startFunc,
			OnStoppedLeading: stopFunc,
		},
	})
	return l, err
}
