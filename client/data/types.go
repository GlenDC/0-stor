package data

import (
	"errors"
	"runtime"
)

var (
	// ErrInvalidResult is an error returned
	// when data was received from the zstordb server,
	// but it wasn't valid.
	ErrInvalidResult = errors.New("invalid result received")
)

var (
	// DefaultJobCount is the default job count that is used
	// for a data client, for those (iterator) methods,
	// that support fetching over multiple goroutines at once.
	DefaultJobCount = runtime.NumCPU()
)

type (
	// Namespace contains information about a namespace.
	// None of this information is directly stored somewhere,
	// and instead it is gathered upon request.
	Namespace struct {
		Label               string
		ReadRequestPerHour  int64
		WriteRequestPerHour int64
		NrObjects           int64
	}

	// Object contains the information stored for an object.
	// The Data and ReferenceList are stored separately,
	// but are composed together in this data structure upon request.
	Object struct {
		Key           []byte
		Data          []byte
		ReferenceList []string
	}

	// ObjectKeyResult is the (stream) data type,
	// used as the result data type, when fetching the keys
	// of all objects stored in the current namespace.
	//
	// Only in case of an error, the Error property will be set,
	// in all other cases only the Key property will be set.
	ObjectKeyResult struct {
		Key   []byte
		Error error
	}
)

// ObjectStatus defines the status of an object,
// it can be retrieved using the Check Method of the Client API.
type ObjectStatus uint8

// ObjectStatus enumeration values.
const (
	// The Object is missing.
	ObjectStatusMissing ObjectStatus = iota
	// The Object is OK.
	ObjectStatusOK
	// The Object is corrupted.
	ObjectStatusCorrupted
)

// String implements Stringer.String
func (status ObjectStatus) String() string {
	return _ObjectStatusEnumToStringMapping[status]
}

// private constants for the string

const _ObjectStatusStrings = "missingokcorrupted"

var _ObjectStatusEnumToStringMapping = map[ObjectStatus]string{
	ObjectStatusMissing:   _ObjectStatusStrings[:7],
	ObjectStatusOK:        _ObjectStatusStrings[7:9],
	ObjectStatusCorrupted: _ObjectStatusStrings[9:],
}
