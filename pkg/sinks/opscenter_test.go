package sinks

import (
	"context"
	"reflect"
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

func newMockedCreateOps(id string) *mockedCreateOps {
	return &mockedCreateOps{
		Resp: ssm.CreateOpsItemOutput{OpsItemId: aws.String(id)},
	}

}

func (m *mockedCreateOps) CreateOpsItemWithContext(ctx aws.Context, in *ssm.CreateOpsItemInput, o ...request.Option) (*ssm.CreateOpsItemOutput, error) {
	m.Input = *in
	return &m.Resp, nil
}

func (m *mockedCreateOps) GetInput() ssm.CreateOpsItemInput {
	return m.Input
}

func makeNotifications(input ...string) []*ssm.OpsItemNotification {
	ns := make([]*ssm.OpsItemNotification, 0)
	for _, v := range input {
		ns = append(ns, &ssm.OpsItemNotification{Arn: aws.String(v)})
	}
	return ns
}

func makeRelatedOpsItems(input ...string) []*ssm.RelatedOpsItem {
	ris := make([]*ssm.RelatedOpsItem, 0)
	for _, v := range input {
		ris = append(ris, &ssm.RelatedOpsItem{OpsItemId: aws.String(v)})
	}
	return ris
}

func makeTags(input map[string]string) []*ssm.Tag {
	tvs := make([]*ssm.Tag, 0)
	for k, v := range input {
		tvs = append(tvs, &ssm.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return tvs
}
func makeOperationalData(input map[string]string) map[string]*ssm.OpsItemDataValue {
	oids := make(map[string]*ssm.OpsItemDataValue)
	for k, v := range input {
		oids[k] = &ssm.OpsItemDataValue{Type: aws.String("SearchableString"), Value: aws.String(v)}
	}
	return oids
}

func TestOpsCenterSink_Send(t *testing.T) {
	m := newMockedCreateOps("id123456")
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
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantInput ssm.CreateOpsItemInput
	}{
		{"Simple Create", fields{
			&OpsCenterConfig{
				Title:           "{{ .Message }}",
				Category:        "{{ .Reason }}",
				Description:     "Event {{ .Reason }} for {{ .InvolvedObject.Namespace }}/{{ .InvolvedObject.Name }} on K8s cluster",
				Notifications:   []string{"sns1", "sns2"},
				OperationalData: map[string]string{"Reason": "{{ .Reason }}"},
				Priority:        "6",
				Region:          "us-east1",
				RelatedOpsItems: []string{"ops1", "ops2"},
				Severity:        "6",
				Source:          "production",
				Tags:            map[string]string{"ENV": "{{ .InvolvedObject.Namespace }}"},
			},
			m,
		}, args{context.Background(), ev}, false,
			ssm.CreateOpsItemInput{
				Category:        aws.String("my reason"),
				Description:     aws.String("Event my reason for prod/nginx-server-123abc-456def on K8s cluster"),
				Notifications:   makeNotifications("sns1", "sns2"),
				OperationalData: makeOperationalData(map[string]string{"Reason": "my reason"}),
				Priority:        aws.Int64(6),
				RelatedOpsItems: makeRelatedOpsItems("my reason"),
				Severity:        aws.String("6"),
				Source:          aws.String("production"),
				Tags:            makeTags(map[string]string{"ENV": "prod"}),
				Title:           aws.String("Successfully pulled image \"nginx:latest\""),
			},
		},
		{"Invalid Priority: Want err", fields{
			&OpsCenterConfig{
				Title:           "{{ .Message }}",
				Category:        "{{ .Reason }}",
				Description:     "Event {{ .Reason }} for {{ .InvolvedObject.Namespace }}/{{ .InvolvedObject.Name }} on K8s cluster",
				Notifications:   []string{"sns1", "sns2"},
				OperationalData: map[string]string{"Reason": "{{ .Reason }}"},
				Priority:        "asdf",
				Region:          "us-east1",
				RelatedOpsItems: []string{"ops1", "ops2"},
				Severity:        "6",
				Source:          "production",
				Tags:            map[string]string{"ENV": "{{ .InvolvedObject.Namespace }}"},
			},
			m,
		}, args{context.Background(), ev}, true,
			ssm.CreateOpsItemInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OpsCenterSink{
				cfg: tt.fields.cfg,
				svc: tt.fields.svc,
			}
			if err := s.Send(tt.args.ctx, tt.args.ev); (err != nil) != tt.wantErr {
				t.Errorf("OpsCenterSink.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m.Input, tt.wantInput) && tt.wantErr != true {
				t.Errorf("OpsCenterSink.Send()  \nReturned:\n%v, \nWanted:\n %v", m.Input, tt.wantInput)
			}
		})
	}
}
