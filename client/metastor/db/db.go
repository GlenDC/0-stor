package db

import "errors"

var (
	// ErrNotFound is the error returned by metadata KV database,
	// in case metadata requested couldn't be found.
	ErrNotFound = errors.New("key couldn't be found")
)

// DB interface is the interface defining how to interact with a key value store,
// as ued for metadata storage. ALl DB implements are assumed to be threadsafe.
type DB interface {
	// Set given key in the database equal to the processed metadata.
	Set(key, metadata []byte) error
	// Get the stored metadata from the database using the given key.
	Get(key []byte) (metadata []byte, err error)
	// Delete the metadata which is stored as the given key.
	Delete(key []byte) error
	// Update metadata stored as the given key,
	// as an in-memory-transaction, providing protection against data races.
	// When wishing to update metadata always use this method,
	// rather than a combination of Set+Get.
	Update(key []byte, cb UpdateCallback) error

	// Close any open (database) resources.
	Close() error
}

// UpdateCallback is the type of callback used to update the processed (encoded)
// metadata, which was already stored, previously.
type UpdateCallback func(orgMetadata []byte) (newMetadata []byte, err error)
