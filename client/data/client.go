package data

import "context"

// Client defines the API for any client,
// used to interface with a zstordb server.
// It allows you to manage objects,
// as well as get information about them and their namespaces.
//
// All operations work within a namespace,
// which is defined by the label given when creating
// this  client.
//
// If the server requires authentication,
// this will have to be configured when creating the client as well, otherwise the methods of this interface will fail.
type Client interface {
	// Set an object, either overwriting an existing key,
	// or creating a new one.
	SetObject(object Object) error

	// Get an existing object, linked to a given key.
	GetObject(key []byte) (*Object, error)

	// Get multiple existing objects,
	// linked to the given keys.
	GetObjectIterator(ctx context.Context, keys [][]byte) (<-chan ObjectResult, error)

	// ExistObject returns whether or not an object,
	// for a given key, exists.
	ExistObject(key []byte) (bool, error)

	// ExistObject returns whether or not
	// one or multiple objects, or (a) given key(s), exists.
	ExistObjectIterator(ctx context.Context, keys [][]byte) (<-chan ObjectExistResult, error)

	// DeleteObjects deletes multiple objects,
	// using the given keys.
	// Deleting an non-existing object is considered valid.
	DeleteObjects(keys ...[]byte) error

	// GetObjectStatus returns the status of an object,
	// indicating whether it's OK, missing or corrupt.
	GetObjectStatus(key []byte) (ObjectStatus, error)

	// GetObjectStatusIterator returns an iterator,
	// which will return a status of
	// all objects for which a key was given in the request,
	// indicating for each of them whether they're
	// OK, missing or corrupt.
	//
	// For each returned result the key of the object is returned as well.
	//
	// In case an error while the iterator is active,
	// it will be returned as part of the last returned result,
	// which is then considered to be invalid.
	// When an error is returned, as part of a result,
	// the iterator channel will be automatically closed as soon
	// as that item is received.
	GetObjectStatusIterator(ctx context.Context, keys [][]byte) (<-chan ObjectStatusResult, error)

	// ListObjectKeyIterator returns an iterator,
	// from which the keys of all stored objects within the namespace
	// (identified by the given label), an be retrieved.
	//
	// In case an error while the iterator is active,
	// it will be returned as part of the last returned result,
	// which is then considered to be invalid.
	// When an error is returned, as part of a result,
	// the iterator channel will be automatically closed as soon
	// as that item is received.
	ListObjectKeyIterator(ctx context.Context) (<-chan ObjectKeyResult, error)

	// GetNamespace returns the available information of a namespace.
	GetNamespace() (*Namespace, error)

	// SetReferenceList allows you to create a new reference list
	// or overwrite an existing reference list,
	// for a given object.
	// If refList is nil, the stored refList will be deleted.
	SetReferenceList(key []byte, refList []string) error

	// GetReferenceList returns an existing reference list
	// for a given object. An error is returned in case no
	// reference exists for that object.
	GetReferenceList(key []byte) ([]string, error)

	// AppendReferenceList appends the given references
	// to the end of the reference list of the given object.
	// If no reference list existed for this object prior to this call,
	// this method will behave the same as SetReferenceList.
	AppendReferenceList(key []byte, refList []string) error

	// DeleteReferenceList removes the references of the given list,
	// from the references of the existing list.
	// It also returns all references which couldn't be deleted
	// from the existing list, because these references did not
	// exist in the given reference list.
	DeleteReferenceList(key []byte, refList []string) error

	// Close any open resources for this data client.
	Close() error
}
