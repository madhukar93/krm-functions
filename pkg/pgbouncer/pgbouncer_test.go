package pgbouncer

import (
	"fmt"
	"testing"

	esapi "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Function to parse the string to ExternalSecret
func parseStringToExternalSecret(input string) *esapi.ExternalSecret {
	secret := esapi.ExternalSecret{}
	err := yaml.Unmarshal([]byte(input), &secret)
	if err != nil {
		fmt.Println(err)
	}
	return &secret
}

// Below is table driven test for validateConnectionSecret function
// This test is used to validate the connection secret
// It checks if the connection secret is nil or empty
// It checks if the connection secret is missing any of the fields
// It checks if the connection secret is valid
func TestValidateConnectionSecret(t *testing.T) {
	var tests = []struct {
		name     string
		secret   *esapi.ExternalSecret
		expected error
	}{
		// Test case: If all fields are present
		{
			name:     "all fields present",
			secret:   parseStringToExternalSecret(allPresent),
			expected: nil,
		},
		// Test case: If connection secret is empty
		{
			name:     "empty secret",
			secret:   &esapi.ExternalSecret{},
			expected: fmt.Errorf("Secret is empty"),
		},
		// Test case: If connection secret is missing one of the fields
		{
			name:     "missing field POSTGRESQL_DATABASE",
			secret:   parseStringToExternalSecret(missingDatabase),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_DATABASE"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateConnectionSecret(test.secret)
			if err != test.expected {
				t.Errorf("unexpected error: got %v, want %v", err, test.expected)
			}
		})
	}
}
