# kubernetes-event-exporter

> This tool is presented at [KubeCon 2019 San Diego](https://kccncna19.sched.com/event/6aa61eca397e4ff2bdbb2845e5aebb81).

This tool allows exporting the often missed Kubernetes events to various outputs so that they can be used for
observability or alerting purposes. You won't believe what you are missing.

## Deployment

Head on to `deploy/` folder and apply the YAMLs in the given filename order. Do not forget to modify the
`deploy/01-config.yaml` file to your configuration needs. The additional information for configuration is as follows:

## Configuration

Configuration is done via a YAML file, when run in Kubernetes, ConfigMap. The tool watches all the events and
user has to option to filter out some events, according to their properties. Critical events can be routed to alerting
tools such as Opsgenie, or all events can be dumped to an Elasticsearch instance. You can use namespaces, labels on the
related object to route some Pod related events to owners via Slack. The final routing is a tree which allows
flexibility. It generally looks like following:

```yaml
route:
  # Main route
  routes:
    # This route allows dumping all events because it has no fields to match and no drop rules.
    - match:
        - receiver: dump
    # This starts another route, drops all the events in *test* namespaces and Normal events
    # for capturing critical events
    - drop:
        - namespace: "*test*"
        - type: "Normal"
      match:
        - receiver: "critical-events-queue"
    # This a final route for user messages
    - match:
        - kind: "Pod|Deployment|ReplicaSet"
          labels:
            version: "dev"
          receiver: "slack"
receivers:
# See below for configuring the receivers
```

* A `match` rule is exclusive, all conditions must be matched to the event.
* During processing a route, `drop` rules are executed first to filter out events.
* The `match` rules in a route are independent of each other. If an event matches a rule, it goes down it's subtree.
* If all the `match` rules are matched, the event is passed to the `receiver`.
* A route can have many sub-routes, forming a tree.
* Routing starts from the root route.

### Opsgenie

[Opsgenie](https://www.opsgenie.com) is an alerting and on-call management tool. kubernetes-event-exporter can push to
events to Opsgenie so that you can notify the on-call when something critical happens. Alerting should be precise and
actionable, so you should carefully design what kind of alerts you would like in Opsgenie. A good starting point might
be filtering out Normal type of events, while some additional filtering can help. Below is an example configuration.

```yaml
# ...
receivers:
  - name: "alerts"
    opsgenie:
      apiKey: xxx
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

### Webhooks/HTTP

Webhooks are the easiest way of integrating this tool to external systems. It allows templating & custom headers which
allows you to push events to many possible sources out there. See [Customizing Payload] for more information.

```yaml
# ...
receivers:
  - name: "alerts"
    webhook:
      endpoint: "https://my-super-secret-service.com"
      headers:
        X-API-KEY: "123"
        User-Agent: kube-event-exporter 1.0
      layout: # Optional
```

### Elasticsearch

[Elasticsearch](https://www.elastic.co/) is a full-text, distributed search engine which can also do powerful
aggregations. You may decide to push all events to Elasticsearch and do some interesting queries over time to find out
which images are pulled, how often pod schedules happen etc. You
can [watch the presentation](https://static.sched.com/hosted_files/kccncna19/d0/Exporting%20K8s%20Events.pdf)
in Kubecon to see what else you can do with aggregation and reporting.

```yaml
# ...
receivers:
  - name: "dump"
    elasticsearch:
      hosts:
        - http://localhost:9200
      index: kube-events
      # Ca be used optionally for time based indices, accepts Go time formatting directives
      indexFormat: "kube-events-{2006-01-02}"
      username: # optional
      password: # optional
      cloudID: # optional
      apiKey: # optional
      # If set to true, it allows updating the same document in ES (might be useful handling count)
      useEventID: true|false
      # Type should be only used for clusters Version 6 and lower.
      # type: kube-event
      # If set to true, all dots in labels and annotation keys are replaced by underscores. Defaults false
      deDot: true|false
      layout: # Optional
      tls: # optional, advanced options for tls
        insecureSkipVerify: true|false # optional, if set to true, the tls cert won't be verified
        serverName: # optional, the domain, the certificate was issued for, in case it doesn't match the hostname used for the connection
        caFile: # optional, path to the CA file of the trusted authority the cert was signed with 
```
### OpenSsearch

[OpenSearch](https://opensearch.org/) is a community-driven, open source search and analytics suite derived from Apache 2.0 licensed Elasticsearch 7.10.2 & Kibana 7.10.2.
OpenSearch enables people to easily ingest, secure, search, aggregate, view, and analyze data. These capabilities are popular for use cases such as application search, log analytics, and more.
You may decide to push all events to OpenSearch and do some interesting queries over time to find out
which images are pulled, how often pod schedules happen etc.

```yaml
# ...
receivers:
  - name: "dump"
    opensearch:
      hosts:
        - http://localhost:9200
      index: kube-events
      # Ca be used optionally for time based indices, accepts Go time formatting directives
      indexFormat: "kube-events-{2006-01-02}"
      username: # optional
      password: # optional
      # If set to true, it allows updating the same document in ES (might be useful handling count)
      useEventID: true|false
      # Type should be only used for clusters Version 6 and lower.
      # type: kube-event
      # If set to true, all dots in labels and annotation keys are replaced by underscores. Defaults false
      deDot: true|false
      layout: # Optional
      tls: # optional, advanced options for tls
        insecureSkipVerify: true|false # optional, if set to true, the tls cert won't be verified
        serverName: # optional, the domain, the certificate was issued for, in case it doesn't match the hostname used for the connection
        caFile: # optional, path to the CA file of the trusted authority the cert was signed with 
```

### Slack

Slack is a cloud-based instant messaging platform where many people use it for integrations and getting notified by
software such as Jira, Opsgenie, Google Calendar etc. and even some implement ChatOps on it. This tool also allows
exporting events to Slack channels or direct messages to persons. If your objects in Kubernetes, such as Pods,
Deployments have real owners, you can opt-in to notify them via important events by using the labels of the objects. If
a Pod sandbox changes and it's restarted, or it cannot find the Docker image, you can immediately notify the owner.

```yaml
# ...
receivers:
  - name: "slack"
    slack:
      token: YOUR-API-TOKEN-HERE
      channel: "@{{ .InvolvedObject.Labels.owner }}"
      message: "{{ .Message }}"
      color: # optional
      title: # optional
      author_name: # optional
      footer: # optional
      fields:
        namespace: "{{ .Namespace }}"
        reason: "{{ .Reason }}"
        object: "{{ .Namespace }}"

```

### Kinesis

Kinesis is an AWS service allows to collect high throughput messages and allow it to be used in stream processing.

```yaml
# ...
receivers:
  - name: "kinesis"
    kineis:
      streamName: "events-pipeline"
      region: us-west-2
      layout: # Optional
```

### Firehose

Firehose is an AWS service providing high throughput message collection for use in stream processing.

```yaml
# ...
receivers:
  - name: "firehose"
    firehose:
      deliveryStreamName: "events-pipeline"
      region: us-west-2
      layout: # Optional
```
### SNS

SNS is an AWS service for highly durable pub/sub messaging system.

```yaml
# ...
receivers:
  - name: "sns"
    sns:
      topicARN: "arn:aws:sns:us-east-1:1234567890123456:mytopic"
      region: "us-west-2"
      layout: # Optional
```

### SQS

SQS is an AWS service for message queuing that allows high throughput messaging.

```yaml
# ...
receivers:
  - name: "sqs"
    sqs:
      queueName: "/tmp/dump"
      region: us-west-2
      layout: # Optional
```

### File

For some debugging purposes, you might want to push the events to files. Or you can already have a logging tool that can
ingest these files and it might be a good idea to just use plain old school files as an integration point.

```yaml
# ...
receivers:
  - name: "file"
    file:
      path: "/tmp/dump"
      layout: # Optional
```

### Stdout

Standard out is also another file in Linux. `logLevel` refers to the application logging severity - available levels
`trace`, `debug`, `info`, `warn`, `error`, `fatal` and `panic`. When not specified, default level is set to `info`. You
can use the following configuration as an example.

By default, events emit with eventime > 5seconds since catching are not collected.
You can set this period with trottlePeriod in seconds. Consider to increase time of seconds to catch more events like "Backoff".

```yaml
logLevel: error
logFormat: json
trottlePeriod: 5
route:
  routes:
    - match:
        - receiver: "dump"
receivers:
  - name: "dump"
    stdout: { }
```

### Kafka

Kafka is a popular tool used for real-time data pipelines. You can combine it with other tools for further analysis.

```yaml
receivers:
  - name: "kafka"
    kafka:
      clientId: "kubernetes"
      topic: "kube-event"
      brokers:
        - "localhost:9092"
      compressionCodec: "snappy"
      tls:
        enable: true
        certFile: "kafka-client.crt"
        keyFile: "kafka-client.key"
        caFile: "kafka-ca.crt"
      sasl:
        enable: true
        username: "kube-event-producer"
        passsord: "kube-event-producer-password"
      layout: #optionnal
        kind: {{ .InvolvedObject.Kind }}
        namespace: {{ .InvolvedObject.Namespace }}
        name: {{ .InvolvedObject.Name }}
        reason: {{ .Reason }}
        message: {{ .Message }}
        type: {{ .Type }}
        createdAt: {{ .GetTimestampISO8601 }}
```

### OpsCenter

[OpsCenter](https://docs.aws.amazon.com/systems-manager/latest/userguide/OpsCenter.html) provides a central location
where operations engineers and IT professionals can view, investigate, and resolve operational work items (OpsItems)
related to AWS resources. OpsCenter is designed to reduce mean time to resolution for issues impacting AWS resources.
This Systems Manager capability aggregates and standardizes OpsItems across services while providing contextual
investigation data about each OpsItem, related OpsItems, and related resources. OpsCenter also provides Systems Manager
Automation documents (runbooks) that you can use to quickly resolve issues. You can specify searchable, custom data for
each OpsItem. You can also view automatically-generated summary reports about OpsItems by status and source.

```yaml
# ...
receivers:
  - name: "alerts"
    opscenter:
    title: "{{ .Message }}",
    category: "{{ .Reason }}", # Optional
    description: "Event {{ .Reason }} for {{ .InvolvedObject.Namespace }}/{{ .InvolvedObject.Name }} on K8s cluster",
    notifications: # Optional: SNS ARN
      - "sns1"
      - "sns2"
  operationalData: # Optional
    - Reason: ""{ { .Reason } }"}"
  priority: "6", # Optional
  region: "us-east1",
  relatedOpsItems: # Optional: OpsItems ARN
    - "ops1"
    - "ops2"
    severity: "6" # Optional
    source: "production"
  tags: # Optional
    - ENV: "{{ .InvolvedObject.Namespace }}"
```

### Customizing Payload

Some receivers allow customizing the payload. This can be useful to integrate it to external systems that require the
data be in some format. It is designed to reduce the need for code writing. It allows mapping an event using Go
templates, with [sprig](github.com/Masterminds/sprig) library additions. It supports a recursive map definition, so that
you can create virtually any kind of JSON to be pushed to a webhook, a Kinesis stream, SQS queue etc.

```yaml
# ...
receivers:
  - name: pipe
    kinesis:
      region: us-west-2
      streamName: event-pipeline
      layout:
        region: "us-west-2"
        eventType: "kubernetes-event"
        createdAt: "{{ .GetTimestampMs }}"
        details:
          message: "{{ .Message }}"
          reason: "{{ .Reason }}"
          type: "{{ .Type }}"
          count: "{{ .Count }}"
          kind: "{{ .InvolvedObject.Kind }}"
          name: "{{ .InvolvedObject.Name }}"
          namespace: "{{ .Namespace }}"
          component: "{{ .Source.Component }}"
          host: "{{ .Source.Host }}"
          labels: "{{ toJson .InvolvedObject.Labels}}"
```

### Pubsub

Pub/Sub is a fully-managed real-time messaging service that allows you to send and receive messages between independent
applications.

```yaml
receivers:
  - name: "pubsub"
    pubsub:
      gcloud_project_id: "my-project"
      topic: "kube-event"
      create_topic: False
```

### Teams

Microsoft Teams is your hub for teamwork in Office 365. All your team conversations, files, meetings, and apps live
together in a single shared workspace, and you can take it with you on your favorite mobile device.

```yaml
# ...
receivers:
  - name: "ms_teams"
    teams:
      endpoint: "https://outlook.office.com/webhook/..."
      layout: # Optional
```

### Syslog

Syslog sink support enables to write k8s-events to syslog daemon server over tcp/udp. This can also be consumed by
rsyslog.

```yaml
# ...
receivers:
  - name: "syslog"
    syslog:
      network: "tcp"
      address: "127.0.0.1:11514"
      tag: "k8s.event"

```

# BigQuery

Google's query thing

```yaml
receivers:
  - name: "my-big-query"
    bigquery:
      location:
      project:
      dataset:
      table:
      credentials_path:
      batch_size:
      max_retries:
      interval_seconds:
      timeout_seconds:
```

# Pipe

pipe output directly into some file descriptor

```yaml
receivers:
  - name: "my_pipe"
    pipe:
      path: "/dev/stdout"
```

# AWS EventBridge

```yaml
receivers:
 - name: "eventbridge"
   eventbridge:
     detailType: "deployment"
     source: "cd"
     eventBusName: "default"
     region: "ap-southeast-1"
     details:
       message: "{{ .Message }}"
       namespace: "{{ .Namespace }}"
       reason: "{{ .Reason }}"
       object: "{{ .Namespace }}"

```
