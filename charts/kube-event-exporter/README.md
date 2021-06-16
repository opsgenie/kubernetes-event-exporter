
## Kube Event Exporter

This tool allows exporting the often missed Kubernetes events to various outputs so that they can be used for observability or alerting purposes.

## Prerequisites

- Kubernetes 1.14+
- Helm 3.2+

## Installing the Chart

First, download the git repository.

```shell
$ helm install kubeevent -n monitoring ./charts/kube-event-exporter
```

## Configuration
The following tables lists the configurable parameters of the chart and their default values.

| Parameter                             | Description                                                                                                                                                                                 | Default                         |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------|
| imagePullSecrets    | Image pull secrets.                                                                                                                                                                         | []
| exporter.config    |  Config values for exporter                                                                                                                                                                       | dump to stdout
