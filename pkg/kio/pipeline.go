package kio

import (
	"io"

	"sigs.k8s.io/kustomize/kyaml/kio"
)

// NewPipeline creates a new kio.Pipeline with the given reader, writer, and keepReaderAnnotations flag.
// It returns the created pipeline.
//
// Parameters:
// - reader: an io.Reader used to read the input data
// - writer: an io.Writer used to write the output data
// - keepReaderAnnotations: a boolean flag indicating whether to keep the annotations from the input data
//
// Return:
// A kio.Pipeline object that reads data from the given reader, runs it through a filter, and writes it to the given writer.
//
// Pipeline reads Resource Configuration from a set of Inputs, applies some
// transformation filters, and writes the results to a set of Outputs.
func NewPipeline(reader io.Reader, writer io.Writer, keepReaderAnnotations bool) kio.Pipeline {
	rw := &kio.ByteReadWriter{Reader: reader, Writer: writer, KeepReaderAnnotations: keepReaderAnnotations}
	return kio.Pipeline{
		Inputs:  []kio.Reader{rw},             // read the inputs into a slice
		Filters: []kio.Filter{Filter{rw: rw}}, // run the filter against the inputs
		Outputs: []kio.Writer{rw},             // copy the inputs to the output
	}
}
