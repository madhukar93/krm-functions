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
func TestValidateConnectionSecret(t *testing.T) {
	var tests = []struct {
		name     string
		secret   *esapi.ExternalSecret
		expected error
	}{
		// Test case: If all fields are present in connection secret
		{
			name:     "all fields present",
			secret:   parseStringToExternalSecret(allPresent),
			expected: nil,
		},
		// Test case: If connection secret is missing some fields
		{
			name:     "missing some fields",
			secret:   parseStringToExternalSecret(missingFields),
			expected: fmt.Errorf("Some of the fields are missing from secret. Required fields are: [POSTGRESQL_PASSWORD POSTGRESQL_HOST POSTGRESQL_PORT POSTGRESQL_USERNAME POSTGRESQL_DATABASE]"),
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
