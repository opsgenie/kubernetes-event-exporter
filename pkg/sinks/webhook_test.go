package sinks

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func funcEqual(a, b interface{}) bool {
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()
	return av.InterfaceData() == bv.InterfaceData()
}

func TestNewWebhook(t *testing.T) {

	type args struct {
		cfg *WebhookConfig
	}

	tests := []struct {
		name    string
		args    args
		want    serializeEvent
		wantErr bool
	}{
		{
			"Default Content-Type header",
			args{cfg: &WebhookConfig{Headers: map[string]string{}}},
			serializeEventWithLayout,
			false,
		},
		{
			"JSON Content-Type header",
			args{cfg: &WebhookConfig{
				Headers: map[string]string{
					"Content-Type": "application/json",
				}},
			},
			serializeEventWithLayout,
			false,
		},
		{
			"XML Content-Type header 1",
			args{
				cfg: &WebhookConfig{
					Headers: map[string]string{
						"Content-Type": "application/xml",
					},
				},
			},
			serializeXMLEventWithLayout,
			false,
		},
		{
			"XML Content-Type header 2",
			args{
				cfg: &WebhookConfig{
					Headers: map[string]string{
						"Content-Type": "application/xml;encoding=utf-8",
					},
				},
			},
			serializeXMLEventWithLayout,
			false,
		},
		{
			"XML Content-Type not managed",
			args{
				cfg: &WebhookConfig{
					Headers: map[string]string{
						"Content-Type": "text/html",
					},
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWebhook(tt.args.cfg)
			if fail := err != nil; fail {
				if fail != tt.wantErr {
					t.Errorf("NewWebhook() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				errMsg := "Wrong Content-Type in headers"
				if !(strings.Contains(err.Error(), errMsg)) {
					t.Errorf("NewWebhook() error: %q, wantErrMsg: %q", err.Error(), errMsg)
				}

				// if err, return anyway
				return
			}

			w, ok := got.(*Webhook)
			require.True(t, ok)
			require.True(t, funcEqual(w.serializer, tt.want))
		})
	}
}
