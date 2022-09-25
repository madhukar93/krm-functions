package main

import (
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	// "k8s.io/apimachinery/pkg/util/yaml"
)

/*
TODO
- preserve comments
- create generate/transform semantics in the framework
- try to leverage more of kyaml/fn package
- have validations
*/

func main() {
	rw := &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		OmitReaderAnnotations: true,
		KeepReaderAnnotations: true,
	}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw}, // read the inputs into a slice
		Filters: []kio.Filter{WorkloadsFilter{rw: rw}},
		Outputs: []kio.Writer{rw}, // copy the inputs to the output
	}
	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}
