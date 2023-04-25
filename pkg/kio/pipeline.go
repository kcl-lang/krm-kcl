package kio

import (
	"io"

	"sigs.k8s.io/kustomize/kyaml/kio"
)

func NewPipeline(reader io.Reader, writer io.Writer, keepReaderAnnotations bool) kio.Pipeline {
	rw := &kio.ByteReadWriter{Reader: reader, Writer: writer, KeepReaderAnnotations: keepReaderAnnotations}
	return kio.Pipeline{
		Inputs:  []kio.Reader{rw},             // read the inputs into a slice
		Filters: []kio.Filter{Filter{rw: rw}}, // run the filter against the inputs
		Outputs: []kio.Writer{rw},             // copy the inputs to the output
	}
}
