package client

import (
	"fmt"
)

// ErrNotImplemented is returned when a method is not implemented on the client
var ErrNotImplemented = fmt.Errorf("not implemented")

// ErrVersionNotFound is returned when a client version is not found/supported
var ErrVersionNotFound = fmt.Errorf("client version not found")

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

// GetClientWithVersion returns a client that matching the provided 0-stor version string
func GetClientWithVersion(version string, addr, namespace, jwt string) (Client, error) {
	switch version {
	case "v1.1.0a8":
		return NewV110a8Client(addr, namespace, jwt)
	case "v1.1.0b2":
	default:
		return nil, ErrVersionNotFound
	}
}
