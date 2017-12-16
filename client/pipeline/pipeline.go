package pipeline

import (
	"io"

	"github.com/zero-os/0-stor/client/components/crypto"
	"github.com/zero-os/0-stor/client/metastor"
	"github.com/zero-os/0-stor/client/pipeline/processing"
)

// Pipeline ...TODO: desc
type Pipeline interface {
	Write(r io.Reader, refList []string) (chunks []metastor.Chunk, err error)
	Read(chunks []metastor.Chunk, w io.Writer) (refList []string, err error)
}

type (
	HasherConstructor    func() (crypto.Hasher, error)
	ProcessorConstructor func() (processing.Processor, error)
)
