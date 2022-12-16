// +kubebuilder:object:generate=true
// +groupName=krm
package pgbouncer

import (
	"testing"

	esapi "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
)

func Test_validateConnectionSecret(t *testing.T) {
	type args struct {
		secret *esapi.ExternalSecret
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConnectionSecret(tt.args.secret); (err != nil) != tt.wantErr {
				t.Errorf("validateConnectionSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
