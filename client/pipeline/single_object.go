package pipeline

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/zero-os/0-stor/client/components/storage"
	"github.com/zero-os/0-stor/client/datastor"
	"github.com/zero-os/0-stor/client/metastor"
)

func NewSingleObjectPipeline(hc HasherConstructor, pc ProcessorConstructor, os storage.ObjectStorage) *SingleObjectPipeline {
	if hc == nil {
		panic("no HasherConstructor given")
	}
	if pc == nil {
		panic("no ProcessorConstructor given")
	}
	if os == nil {
		panic("no ObjectStorage given")
	}
	return &SingleObjectPipeline{
		hasher:    hc,
		processor: pc,
		storage:   os,
	}
}

// SingleObjectPipeline ...TODO: add description
type SingleObjectPipeline struct {
	hasher    HasherConstructor
	processor ProcessorConstructor
	storage   storage.ObjectStorage
}

// Write implements Pipeline.Write
//
// The following graph visualizes the logic of this pipeline's Write method:
//
// +-----------------------------------------------------------------------+
// | io.Reader+Hasher +-> Processor.Write +-> Storage.Write +-> meta.Meta  |
// +-----------------------------------------------------------------------+
//
// As you can see, it is all blocking, sequential and the input data is not split into chunks.
// Meaning this pipeline will always return single chunk, as long as the data was written successfully.
//
// When an error is returned by a sub-call, at any point,
// the function will return immediately with that error.
func (sop *SingleObjectPipeline) Write(r io.Reader, refList []string) ([]metastor.Chunk, error) {
	if r == nil {
		return nil, errors.New("no reader given to read from")
	}

	// create the hasher and processor
	hasher, err := sop.hasher()
	if err != nil {
		return nil, err
	}
	processor, err := sop.processor()
	if err != nil {
		return nil, err
	}

	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	key := hasher.HashBytes(input)

	data, err := processor.WriteProcess(input)
	if err != nil {
		return nil, err
	}

	cfg, err := sop.storage.Write(datastor.Object{
		Key:           key,
		Data:          data,
		ReferenceList: refList,
	})
	if err != nil {
		return nil, err
	}

	return []metastor.Chunk{
		metastor.Chunk{
			Key:    cfg.Key,
			Shards: cfg.Shards,
			Size:   int64(cfg.DataSize),
		},
	}, nil
}

// Read implements Pipeline.Read
//
// The following graph visualizes the logic of this pipeline's Read method:
//
//    +-------------------------------------------------------------+
//    |                                    +----------------------+ |
//    | metastor.Chunk +-> storage.Read +--> Processor.Read +     | |
//    |                                    | Hash/Data Validation | |
//    |                                    +---------------------+  |
//    |                                                |            |
//    |                                io.Writer <-----+            |
//    +-------------------------------------------------------------+
//
// As you can see, it is all blocking, sequential and the input data is expected to be only 1 chunk.
// If less or more than one chunk is given, an error will be returned before the pipeline even starts reading.
//
// When an error is returned by a sub-call, at any point,
// the function will return immediately with that error.
func (sop *SingleObjectPipeline) Read(chunks []metastor.Chunk, w io.Writer) (refList []string, err error) {
	if len(chunks) != 1 {
		return nil, errors.New("unexpected chunk count, SingleObjectPipeline requires one and only one chunk")
	}
	if w == nil {
		return nil, errors.New("nil writer")
	}

	// create the hasher and processor
	hasher, err := sop.hasher()
	if err != nil {
		return nil, err
	}
	processor, err := sop.processor()
	if err != nil {
		return nil, err
	}

	obj, err := sop.storage.Read(storage.ObjectConfig{
		Key:      chunks[0].Key,
		Shards:   chunks[0].Shards,
		DataSize: int(chunks[0].Size),
	})
	if err != nil {
		return nil, err
	}

	data, err := processor.ReadProcess(obj.Data)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(obj.Key, hasher.HashBytes(data)) != 0 {
		return nil, errors.New("object chunk's data and key do not match")
	}

	_, err = w.Write(data)
	return obj.ReferenceList, err
}

var (
	_ Pipeline = (*SingleObjectPipeline)(nil)
)
