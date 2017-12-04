package client

import "fmt"

// ErrNotImplemented is the error return when a method is not implemented on the client
var ErrNotImplemented = fmt.Errorf("not implemented")

// Client defines a grpc benchmarking client
type Client interface {
	// ObjectCreate creates an object
	ObjectCreate(id, data []byte, refList []string) error

	// ObjectGet retrieve object from the store
	ObjectGet(id []byte) (*Object, error)
}

// Object represents an Object that is sent and received to and from the 0-stor server
type Object interface {
	GetKey() []byte
	GetValue() []byte
	GetReferenceList() []string
}
