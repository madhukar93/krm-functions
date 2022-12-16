// +kubebuilder:object:generate=true
// +groupName=krm
package pgbouncer

import (
	"fmt"
	"testing"

	esapi "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"gopkg.in/yaml.v2"
)

// Input test data for connection secret where all the fields are present
var inputSecretTest1 = `
POSTGRESQL_HOST:     "localhost",
POSTGRESQL_PORT:     "5432",
POSTGRESQL_USERNAME: "user",
POSTGRESQL_PASSWORD: "password",
POSTGRESQL_DATABASE: "database",
`

// Input test data for connection secret where POSTGRESQL_DATABASE is missing
var inputSecretTest2 = `
POSTGRESQL_HOST:     "localhost",
POSTGRESQL_PORT:     "5432",
POSTGRESQL_USERNAME: "user",
POSTGRESQL_PASSWORD: "password",
`

// Input test data for connection secret where POSTGRESQL_PASSWORD is missing
var inputSecretTest3 = `
POSTGRESQL_HOST:     "localhost",
POSTGRESQL_PORT:     "5432",
POSTGRESQL_USERNAME: "user",
POSTGRESQL_DATABASE: "database",
`

// Input test data for connection secret where POSTGRESQL_USERNAME is missing
var inputSecretTest4 = `
POSTGRESQL_HOST:     "localhost",
POSTGRESQL_PORT:     "5432",
POSTGRESQL_PASSWORD: "password",
POSTGRESQL_DATABASE: "database",
`

// Input test data for connection secret where POSTGRESQL_PORT is missing
var inputSecretTest5 = `
POSTGRESQL_HOST:     "localhost",
POSTGRESQL_USERNAME: "user",
POSTGRESQL_PASSWORD: "password",
POSTGRESQL_DATABASE: "database",
`

// Input test data for connection secret where POSTGRESQL_HOST is missing
var inputSecretTest6 = `
POSTGRESQL_PORT:     "5432",
POSTGRESQL_USERNAME: "user",
POSTGRESQL_PASSWORD: "password",
POSTGRESQL_DATABASE: "database",
`

// Function to parse the string to ExternalSecret
func parseStringToExternalSecret(input string) *esapi.ExternalSecret {
	var secret esapi.ExternalSecret
	err := yaml.Unmarshal([]byte(input), &secret)
	if err != nil {
		panic(err)
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
		{
			name:     "valid secret",
			secret:   parseStringToExternalSecret(inputSecretTest1),
			expected: nil,
		},
		{
			name:     "nil secret",
			secret:   nil,
			expected: fmt.Errorf("ConnectionSecret is empty"),
		},
		{
			name:     "empty secret",
			secret:   &esapi.ExternalSecret{},
			expected: fmt.Errorf("All fields in ConnectionSecret is missing"),
		},
		{
			name:     "missing field",
			secret:   parseStringToExternalSecret(inputSecretTest2),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_DATABASE"),
		},
		{
			name:     "missing field POSTGRESQL_PASSWORD",
			secret:   parseStringToExternalSecret(inputSecretTest3),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_PASSWORD"),
		},
		{
			name:     "missing field POSTGRESQL_USERNAME",
			secret:   parseStringToExternalSecret(inputSecretTest4),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_USERNAME"),
		},
		{
			name:     "missing field POSTGRESQL_PORT",
			secret:   parseStringToExternalSecret(inputSecretTest5),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_PORT"),
		},
		{
			name:     "missing field POSTGRESQL_HOST",
			secret:   parseStringToExternalSecret(inputSecretTest6),
			expected: fmt.Errorf("ConnectionSecret is missing field POSTGRESQL_HOST"),
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
