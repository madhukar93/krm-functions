package main

import (
	"bytes"
	"testing"

	utils "github.com/bukukasio/krm-functions/pkg/testing"
)

func Test_command(t *testing.T) {
	fnConfigPath := "example/pgbouncer.yaml"

	// TODO
	expected := []byte(`---
	`)

	cmd := cmd()
	cmd.SetArgs([]string{fnConfigPath})
	outbuf := &bytes.Buffer{}
	cmd.SetOut(outbuf)
	if err := cmd.Execute(); err != nil {
		t.Errorf("function failed: %v", err)
	}
	t.Log(outbuf.String())
	if diff, err := utils.YamlDiff(outbuf.Bytes(), expected); err != nil {
		t.Errorf("failed to diff: %v", err)
	} else if diff.String() != "" {
		t.Errorf("Expected output diff: %v", diff.String())
	}
}
