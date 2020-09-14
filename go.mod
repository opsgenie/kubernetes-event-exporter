module github.com/opsgenie/kubernetes-event-exporter

go 1.14

require (
	cloud.google.com/go/bigquery v1.9.0
	cloud.google.com/go/pubsub v1.3.1
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Shopify/sarama v1.24.1
	github.com/aws/aws-sdk-go v1.30.10
	github.com/elastic/go-elasticsearch/v7 v7.4.1
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/klauspost/cpuid v1.2.3 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/nlopes/slack v0.6.0
	github.com/opsgenie/opsgenie-go-sdk-v2 v1.0.3
	github.com/rs/zerolog v1.16.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf // indirect
	google.golang.org/api v0.28.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)
