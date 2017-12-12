package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/zero-os/0-stor/client/datastor"
	"github.com/zero-os/0-stor/server"
)

// NewClient creates a new (in-memory) datastor Client.
// See Client within this package for more information.
func NewClient() *Client {
	return &Client{
		objects: make(map[string]datastor.Object),
	}
}

// Client defines a data client,
// using a very simple golang-map in-memory approach.
//
// This client is only meant for development and testing purposes,
// as it does not connect to an actual zstordb server.
type Client struct {
	objects map[string]datastor.Object
	mux     sync.RWMutex
}

// some private errors returned by the Memory datastor client
var (
	errNilKey     = errors.New("nil key given")
	errNilData    = errors.New("nil data given")
	errNilRefList = errors.New("nil refList given")
)

// SetObject implements datastor.Client.SetObject
func (client *Client) SetObject(object datastor.Object) error {
	if len(object.Key) == 0 {
		return errNilKey
	}
	if len(object.Data) == 0 {
		return errNilData
	}

	obj := object
	obj.Key = make([]byte, len(object.Key))
	copy(obj.Key, object.Key)
	obj.Data = make([]byte, len(object.Data))
	copy(obj.Data, object.Data)
	obj.ReferenceList = make([]string, len(object.ReferenceList))
	copy(obj.ReferenceList, object.ReferenceList)

	client.mux.Lock()
	client.objects[string(obj.Key)] = obj
	client.mux.Unlock()
	return nil
}

// GetObject implements datastor.Client.GetObject
func (client *Client) GetObject(key []byte) (*datastor.Object, error) {
	if len(key) == 0 {
		return nil, errNilKey
	}
	client.mux.RLock()
	object, ok := client.objects[string(key)]
	client.mux.RUnlock()
	if !ok {
		return nil, datastor.ErrKeyNotFound
	}
	return &object, nil
}

// DeleteObject implements datastor.Client.DeleteObject
func (client *Client) DeleteObject(key []byte) error {
	if len(key) == 0 {
		return errNilKey
	}

	client.mux.Lock()
	delete(client.objects, string(key))
	client.mux.Unlock()
	return nil
}

// GetObjectStatus implements datastor.Client.GetObjectStatus
func (client *Client) GetObjectStatus(key []byte) (datastor.ObjectStatus, error) {
	if len(key) == 0 {
		return datastor.ObjectStatus(0), errNilKey
	}

	client.mux.RLock()
	_, ok := client.objects[string(key)]
	client.mux.RUnlock()

	if !ok {
		return datastor.ObjectStatusMissing, nil
	}
	return datastor.ObjectStatusOK, nil
}

// ExistObject implements datastor.Client.ExistObject
func (client *Client) ExistObject(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, errNilKey
	}

	client.mux.RLock()
	_, exists := client.objects[string(key)]
	client.mux.RUnlock()
	return exists, nil
}

// ListObjectKeyIterator implements datastor.Client.ListObjectKeyIterator
func (client *Client) ListObjectKeyIterator(ctx context.Context) (<-chan datastor.ObjectKeyResult, error) {
	client.mux.RLock()
	if len(client.objects) == 0 {
		client.mux.RUnlock()
		return nil, errors.New("no objects available")
	}

	ch := make(chan datastor.ObjectKeyResult, 1)
	go func() {
		defer func() {
			client.mux.RUnlock()
			close(ch)
		}()
		for key := range client.objects {
			select {
			case ch <- datastor.ObjectKeyResult{Key: []byte(key)}:
			case <-ctx.Done():
				return // done
			}
		}
	}()
	return ch, nil
}

// GetNamespace implements datastor.Client.GetNamespace
func (client *Client) GetNamespace() (*datastor.Namespace, error) {
	panic("method not implemented")
}

// SetReferenceList implements datastor.Client.SetReferenceList
func (client *Client) SetReferenceList(key []byte, refList []string) error {
	if len(key) == 0 {
		return errNilKey
	}
	if len(refList) == 0 {
		return errNilRefList
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	object := client.objects[string(key)]
	object.ReferenceList = make([]string, len(refList))
	copy(object.ReferenceList, refList)
	client.objects[string(key)] = object
	return nil
}

// GetReferenceList implements datastor.Client.GetReferenceList
func (client *Client) GetReferenceList(key []byte) ([]string, error) {
	if len(key) == 0 {
		return nil, errNilKey
	}

	client.mux.RLock()
	object := client.objects[string(key)]
	client.mux.RUnlock()

	if len(object.ReferenceList) == 0 {
		return nil, datastor.ErrKeyNotFound
	}
	return object.ReferenceList, nil
}

// GetReferenceCount implements datastor.Client.GetReferenceCount
func (client *Client) GetReferenceCount(key []byte) (int64, error) {
	if len(key) == 0 {
		return 0, errNilKey
	}

	client.mux.RLock()
	object := client.objects[string(key)]
	client.mux.RUnlock()
	return int64(len(object.ReferenceList)), nil
}

// AppendToReferenceList implements datastor.Client.AppendToReferenceList
func (client *Client) AppendToReferenceList(key []byte, refList []string) error {
	if len(key) == 0 {
		return errNilKey
	}
	if len(refList) == 0 {
		return errNilRefList
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	object := client.objects[string(key)]
	list := server.ReferenceList(object.ReferenceList)
	list.AppendReferences(server.ReferenceList(refList))
	object.ReferenceList = []string(list)
	client.objects[string(key)] = object

	return nil
}

// DeleteFromReferenceList implements datastor.Client.DeleteFromReferenceList
func (client *Client) DeleteFromReferenceList(key []byte, refList []string) (int64, error) {
	if len(key) == 0 {
		return 0, errNilKey
	}
	if len(refList) == 0 {
		return 0, errNilRefList
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	object := client.objects[string(key)]
	list := server.ReferenceList(object.ReferenceList)
	list.RemoveReferences(server.ReferenceList(refList))
	object.ReferenceList = []string(list)
	client.objects[string(key)] = object

	return int64(len(list)), nil
}

// DeleteReferenceList implements datastor.Client.DeleteReferenceList
func (client *Client) DeleteReferenceList(key []byte) error {
	if len(key) == 0 {
		return errNilKey
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	object, ok := client.objects[string(key)]
	if !ok {
		return nil // nothing to do
	}

	object.ReferenceList = nil
	client.objects[string(key)] = object
	return nil
}

// Close implements datastor.Client.Close
func (client *Client) Close() error {
	client.mux.Lock()
	client.objects = make(map[string]datastor.Object)
	client.mux.Unlock()
	return nil
}

var (
	_ datastor.Client = (*Client)(nil)
)
