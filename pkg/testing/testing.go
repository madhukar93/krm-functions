package testing

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/wI2L/jsondiff"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func Compare(appfn func() error, in string, expected_output string) bool {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			outC <- ""
		}
		outC <- buf.String()
	}()

	tmpfile, err := os.CreateTemp("", "test-input")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // noerrcheck
	if _, err := tmpfile.Write([]byte(in)); err != nil {
		log.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	os.Stdin = tmpfile

	err = appfn()
	if err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	output := <-outC

	diff := getDiff(output, expected_output)
	if diff != nil {
		log.Println(diff)
		return false
	}
	// log.Println(expected_output)
	return true
}

func getDiff(output string, expected_output string) jsondiff.Patch {
	output_json, err := yaml.ToJSON([]byte(output))
	if err != nil {
		log.Fatal(err)
	}
	expected_output_json, err := yaml.ToJSON([]byte(expected_output))
	if err != nil {
		log.Fatal(err)
	}

	diff, _ := jsondiff.CompareJSON(output_json, expected_output_json)
	return diff
}
