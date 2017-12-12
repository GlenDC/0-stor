package storage

import (
	"errors"
	"runtime"

	"github.com/zero-os/0-stor/client/datastor"
)

var (
	// DefaultJobCount is the default job count used if the API
	// was created with a job count of 0.
	DefaultJobCount = runtime.NumCPU() * 2
)

// TODO:
// Streamline the `metastor.Chunk` and `datastor.Object` names.

// Errors that can be returned by a storage.
var (
	ErrInsufficientShards    = errors.New("data was written to an insufficient amount of shards")
	ErrUnexpectedShardsCount = errors.New("unexpected shards count")
	ErrShardsUnavailable     = errors.New("(too many?) shards are unavailable")
	ErrNotSupported          = errors.New("method is not supported")
	ErrInvalidDataSize       = errors.New("returned object has invalid data size")
)

type Storage interface {
	Write(object datastor.Object) (StorageConfig, error)
	Read(cfg StorageConfig) (datastor.Object, error)
	Repair(cfg StorageConfig) (StorageConfig, error)

	Close() error
}

type StorageConfig struct {
	Key      []byte
	Shards   []string
	DataSize int
}
