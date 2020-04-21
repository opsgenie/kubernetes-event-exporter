package sinks

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockedCreateOps struct {
	ssmiface.SSMAPI
	Resp  ssm.CreateOpsItemOutput
	Input ssm.CreateOpsItemInput
}

func (m mockedCreateOps) CreateOpsItemWithContext(ctx aws.Context, in *ssm.CreateOpsItemInput, o ...request.Option) (*ssm.CreateOpsItemOutput, error) {
	m.Input = *in
	return &m.Resp, nil
}

func (m mockedCreateOps) GetInput() *ssm.CreateOpsItemInput {
	return &m.Input
}

func TestOpsCenterSink_Send(t *testing.T) {
	m := mockedCreateOps{Resp: ssm.CreateOpsItemOutput{OpsItemId: aws.String("id123456")}}
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	ev.Reason = "my reason"
	ev.Type = "Warning"
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Name = "nginx-server-123abc-456def"
	ev.InvolvedObject.Namespace = "prod"
	ev.Message = "Successfully pulled image \"nginx:latest\""
	ev.FirstTimestamp = v1.Time{Time: time.Now()}
	type fields struct {
		cfg *OpsCenterConfig
		svc ssmiface.SSMAPI
	}
	type args struct {
		ctx context.Context
		ev  *kube.EnhancedEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Simple Create", fields{
			&OpsCenterConfig{
				Title:           "{{ .Message }}",
				Category:        "{{ .Reason }}",
				Description:     "Event {{ .Reason }} for {{ .InvolvedObject.Namespace }}/{{ .InvolvedObject.Name }} on K8s cluster",
				Notifications:   []string{"sns1", "sns2"},
				OperationalData: map[string]string{"Reason": "{{ .Reason }}"},
				//Priority:        "gs",
				Region:          "us-east1",
				RelatedOpsItems: []string{"ops1", "ops2"},
				Severity:        "6",
				Source:          "production",
				Tags:            map[string]string{"ENV": "{{ .InvolvedObject.Namespace }}"},
			},
			m,
		}, args{context.Background(), ev}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OpsCenterSink{
				cfg: tt.fields.cfg,
				svc: tt.fields.svc,
			}
			if err := s.Send(tt.args.ctx, tt.args.ev); (err != nil) != tt.wantErr {
				t.Errorf("OpsCenterSink.Send() error = %v, wantErr %v", err, tt.wantErr)
				t.Log(m.GetInput())
			}
		})
	}
}
