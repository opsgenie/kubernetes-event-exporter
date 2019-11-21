# kubernetes-event-exporter

This tool is presented at [KubeCon 2019 San Diego](https://kccncna19.sched.com/event/6aa61eca397e4ff2bdbb2845e5aebb81).
It allows exporting the often missed Kubernetes events to various outputs so that 
they can be used for observability or alerting purposes.

Objects in Kubernetes, such as Pod, Deployment, Ingress, Service publish events
to indicate status updates or problems. Most of the time, these events are 
overlooked and their 1 hour lifespan might cause missing important updates. 
They are also not searchable and cannot be aggregated.

We are open-sourcing our internal tool for publishing the events in Kubernetes 
to Opsgenie, Slack, Elasticsearch, Webhooks, Kinesis, Pub/Sub. It has a 
configuration language for matching events based on various criteria,
such as the content and the related object’s labels. It also has the capability
to route the events intelligently, inspired by Prometheus Alertmanager.

For instance, you can notify an owner of Pod for runtime OCI failures, 
you can aggregate how many times the images are pulled, how many times
container sandbox changes for various resource labels.

# Configuration

The following is a catch-all and route to a single sink. 

```yaml
route:
  match:
    - receiver: "alerts"
receivers:
  - name: "alerts"
    opsgenie:
      apikey: xxx
      priority: "P3"
      message: "Event {{ .Reason }} for {{ .InvolvedObject.Namespace }}/{{ .InvolvedObject.Name }} on K8s cluster"
      alias: "{{ .UID }}"
      description: "<pre>{{ toPrettyJson . }}</pre>"
      tags:
        - "event"
        - "{{ .Reason }}"
        - "{{ .InvolvedObject.Kind }}"
        - "{{ .InvolvedObject.Name }}"
```

# Sinks

- Opsgenie
- HTTP
- Elasticsearch
- AWS Firehose
- File

Upcoming Sinks:

- Redis
- Kafka
- SNS
- Splunk
- Logstash
- CloudWatch
- GCP PubSub
- BigQuery

# Deployment

Currently there is no official Docker image & Kuberentes Deployment YAMLs right now, it's in WIP.