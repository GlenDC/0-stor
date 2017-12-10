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
//
// Errors that can be returned for all methods:
//
// ErrDatabase in case some internal error with the internal database happened,
// see the zstordb logs for more information about this one.
//
// ErrDatabase in case some internal error with the internal database happened,
// see the zstordb logs for more information about this one.
// Report a bug at http://github.com/zero-os/0-stor/issues when this happens.
//
// ErrNilLabel in case the label wasn't specified as metadata.
// As the label is required for all methods.
//
// ErrNilToken in case the JWT token wasn't specified as metadata,
// while this was expected by the jwt-validated server.
//
// ErrUnimplemented shouldn't happen,
// but will be returned in case the method wasn't recognized by a used middleware.
// Report a bug at http://github.com/zero-os/0-stor/issues when this happens.
//
// ErrPermissionDenied is returned in case the used method (action)
// was not permitted for the given JWT token.
// Contact your admin to get the sufficient rights if this happens.
type Client interface {
	// Set an object, either overwriting an existing key,
	// or creating a new one.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrNilData is returned in case no data was given.
	SetObject(object Object) error

	// Get an existing object, linked to a given key.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrObjectDataCorrupted is returned in case the stored data is corrupted.
	// rpctypes.ErrObjectRefListCorrupted is returned in case the stored refList is corrupted.
	GetObject(key []byte) (*Object, error)

	// DeleteObject deletes an object, using a given key.
	// Deleting an non-existing object is considered valid.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	DeleteObject(key []byte) error

	// GetObjectStatus returns the status of an object,
	// indicating whether it's OK, missing or corrupt.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	GetObjectStatus(key []byte) (ObjectStatus, error)

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
	//
	// rpctypes.ErrKeyNotFound is returned in case no
	// stored namespace exist for the used label.
	GetNamespace() (*Namespace, error)

	// SetReferenceList allows you to create a new reference list
	// or overwrite an existing reference list,
	// for a given object.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrNilRefList is returned in case no ref list was given.
	SetReferenceList(key []byte, refList []string) error

	// GetReferenceList returns an existing reference list
	// for a given object.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrKeyNotFound is returned in case no
	// reference exists for that object.
	// rpctypes.ErrObjectRefListCorrupted is returned in case the stored refList is corrupted.
	GetReferenceList(key []byte) ([]string, error)

	// GetReferenceCount returns the reference count
	// for a given object. If no reference list is given,
	// the object is assumed to have 0 references.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrObjectRefListCorrupted is returned in case the stored refList is corrupted.
	GetReferenceCount(key []byte) (int64, error)

	// AppendToReferenceList appends the given references
	// to the end of the reference list of the given object.
	// If no reference list existed for this object prior to this call,
	// this method will behave the same as SetReferenceList.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrNilRefList is returned in case no ref list was given.
	// rpctypes.ErrObjectRefListCorrupted is returned in case the stored refList is corrupted.
	AppendToReferenceList(key []byte, refList []string) error

	// DeleteFromReferenceList removes the references of the given list,
	// from the references of the existing list.
	// It also returns all references which couldn't be deleted
	// from the existing list, because these references did not
	// exist in the given reference list.
	// The amount of references left over after this operation,
	// is returned.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	// rpctypes.ErrNilRefList is returned in case no ref list was given.
	// rpctypes.ErrObjectRefListCorrupted is returned in case the stored refList is corrupted.
	DeleteFromReferenceList(key []byte, refList []string) (int64, error)

	// DeleteReferenceList deletes the stored reference list for the given (object) key.
	// Deleting a non-existent reference list not considered an error.
	//
	// rpctypes.ErrNilKey is returned in case no key was given.
	DeleteReferenceList(key []byte) error

	// Close any open resources for this data client.
	Close() error
}
