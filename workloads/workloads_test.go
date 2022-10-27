package main

import (
	"bytes"
	"testing"

	utils "github.com/bukukasio/krm-functions/pkg/testing"
)

var rolloutOutput = `apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  creationTimestamp: null
  labels:
    app: foobar-api
    part-of: foobar
  name: foobar-api
spec:
  selector:
    matchLabels:
      app: foobar-api
      part-of: foobar
  strategy:
    canary:
      analysis:
        args:
        - name: service-name
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: env
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['app.tokko.io/env']
        - name: version
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['app.tokko.io/version']
        - name: operation
          value: graphql.execute
        - name: p95latency
          value: 500ms
        - name: errorRPM
          value: "0.1"
        startingStep: 2
        templates:
        - templateName: analysis-datadog-request-errors
        - templateName: analysis-datadog-request-p95-latency
      steps:
      - setWeight: 30
      - pause:
          duration: 300
      - setWeight: 60
      - pause:
          duration: 600
      - setWeight: 100
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: foobar-api
        part-of: foobar
    spec:
      containers:
      - command:
        - python
        - server.py
        envFrom:
        - configMapRef:
            name: foobar-api-config
        - secretRef:
            name: foobar-api-database-secrets
        image: foobar
        name: foobar-api
        resources: {}
status:
  blueGreen: {}
  canary: {}
`

var cronOutput = ``
var jobOutput = ``

func Test_rollout(t *testing.T) {
	fnConfigPath := "example/rollout.yaml"
	// TODO
	expected := []byte(rolloutOutput)

	cmd := cmd()
	cmd.SetArgs([]string{fnConfigPath})
	outbuf := &bytes.Buffer{}
	cmd.SetOut(outbuf)
	if err := cmd.Execute(); err != nil {
		t.Errorf("function failed: %v", err)
	}
	t.Log("output", outbuf.String())
	if diff, err := utils.YamlDiff(outbuf.Bytes(), expected); err != nil {
		t.Errorf("failed to diff: %v", err)
	} else if diff.String() != "" {
		t.Errorf("Expected output diff: %v", diff.String())
	}
}
