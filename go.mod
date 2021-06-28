module github.com/opsgenie/kubernetes-event-exporter

go 1.16

require (
	cloud.google.com/go/bigquery v1.9.0
	cloud.google.com/go/pubsub v1.3.1
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Shopify/sarama v1.24.1
	github.com/aws/aws-sdk-go v1.30.10
	github.com/elastic/go-elasticsearch/v7 v7.4.1
	github.com/hashicorp/golang-lru v0.5.3
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/klauspost/cpuid v1.2.3 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/opsgenie/opsgenie-go-sdk-v2 v1.0.3
	github.com/rs/zerolog v1.16.0
	github.com/slack-go/slack v0.9.1
	github.com/stretchr/testify v1.6.1
	google.golang.org/api v0.28.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
