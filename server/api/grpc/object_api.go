package grpc

import (
	"errors"

	"golang.org/x/sync/errgroup"

	"golang.org/x/net/context"

	"github.com/zero-os/0-stor/server"
	serverAPI "github.com/zero-os/0-stor/server/api"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"github.com/zero-os/0-stor/server/db"
	"github.com/zero-os/0-stor/server/encoding"
)

var _ (pb.ObjectManagerServer) = (*ObjectAPI)(nil)

// ObjectAPI implements pb.ObjectManagerServer
type ObjectAPI struct {
	db       db.DB
	jobCount int
}

// NewObjectAPI returns a new ObjectAPI
func NewObjectAPI(db db.DB, jobs int) *ObjectAPI {
	if db == nil {
		panic("no database given to ObjectAPI")
	}
	if jobs <= 0 {
		jobs = DefaultJobCount
	}

	return &ObjectAPI{
		db:       db,
		jobCount: jobs,
	}
}

// SetObject implements ObjectManagerServer.SetObject
func (api *ObjectAPI) SetObject(ctx context.Context, req *pb.SetObjectRequest) (*pb.SetObjectResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	key := req.GetKey()

	// encode the value and store it
	value := req.GetValue()
	data, err := encoding.EncodeObject(server.Object{Data: value})
	if err != nil {
		return nil, err
	}
	valueKey := db.DataKey([]byte(label), key)
	err = api.db.Set(valueKey, data)
	if err != nil {
		return nil, err
	}

	// either delete the reference list, or set it.
	refListkey := db.ReferenceListKey([]byte(label), key)
	refList := req.GetReferenceList()
	if len(refList) == 0 {
		err = api.db.Delete(refListkey)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = encoding.EncodeReferenceList(server.ReferenceList(refList))
		if err != nil {
			return nil, err
		}
		err = api.db.Set(refListkey, data)
		if err != nil {
			return nil, err
		}
	}

	// return the success reply
	return &pb.SetObjectResponse{}, nil
}

// GetObject implements ObjectManagerServer.GetObject
func (api *ObjectAPI) GetObject(ctx context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	result, err := api.getObject([]byte(label), req.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.GetObjectResponse{
		Key:           result.Key,
		Value:         result.Value,
		Status:        result.Status,
		ReferenceList: result.RefList,
	}, nil
}

// GetObjectStream implements ObjectManagerServer.GetObjectStream
func (api *ObjectAPI) GetObjectStream(req *pb.GetObjectStreamRequest, stream pb.ObjectManager_GetObjectStreamServer) error {
	label, err := extractStringFromContext(stream.Context(), serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return unauthenticatedError(err)
	}

	keys := req.GetKeys()
	keyLength := len(keys)
	if keyLength == 0 {
		return errors.New("no keys given")
	}

	// if only one object is given, simply return the single object
	if keyLength == 1 {
		var result *objectResult
		result, err = api.getObject([]byte(label), keys[0])
		if err != nil {
			return err
		}
		return stream.SendMsg(&pb.GetObjectStreamResponse{
			Key:           result.Key,
			Value:         result.Value,
			Status:        result.Status,
			ReferenceList: result.RefList,
		})
	}

	jobCount := api.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	inputCh := make(chan []byte, jobCount)
	outputCh := make(chan pb.GetObjectStreamResponse, jobCount)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// create an errgroup for all the worker routines,
	// including the input one
	group, ctx := errgroup.WithContext(ctx)

	// start the input goroutine,
	// which simply gives all keys over a channel
	group.Go(func() error {
		// close input channel when this goroutine is finished
		// (either because of an error or because all items have been received)
		defer close(inputCh)

		for _, key := range keys {
			select {
			case inputCh <- key:
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	// start the output goroutine,
	// as we are only allowed to send to the stream on a single goroutine
	// (sending on multiple goroutines at once is not safe according to docs)
	group.Go(func() error {
		// local variables reused for each iteration/item
		var (
			resp            pb.GetObjectStreamResponse
			workerStopCount int
		)

		// loop while we can receive responses,
		// or until the context is done
		for {
			select {
			case <-ctx.Done():
				return nil // early exist -> context is done
			case resp = <-outputCh:
				if resp.Key == nil {
					workerStopCount++
					if workerStopCount == jobCount {
						return nil // we're done!
					}
					continue
				}
			}
			err := stream.Send(&resp)
			if err != nil {
				return err
			}
		}
	})

	// start all the workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// local variables reused for each iteration/item
			var (
				err    error
				open   bool
				key    []byte
				result *objectResult
			)
			for {
				// get the next object key,
				// or return in case the input channel or context has been closed
				select {
				case key, open = <-inputCh:
					if !open {
						select {
						// try to return nil object,
						// as to indicate this worker is finished
						case outputCh <- pb.GetObjectStreamResponse{}:
						case <-ctx.Done():
						}
						return nil
					}
				case <-ctx.Done():
					return nil
				}

				// fetch + decode the object
				result, err = api.getObject([]byte(label), key)
				if err != nil {
					return err
				}

				resp := pb.GetObjectStreamResponse{
					Key:           result.Key,
					Value:         result.Value,
					Status:        result.Status,
					ReferenceList: result.RefList,
				}
				// return the new object
				select {
				case outputCh <- resp:
				case <-ctx.Done():
					// return, context is done
					return nil
				}
			}
		})
	}

	// wait until all contexts are finished
	return group.Wait()
}

// ExistObject implements ObjectManagerServer.ExistObject
func (api *ObjectAPI) ExistObject(ctx context.Context, req *pb.ExistObjectRequest) (*pb.ExistObjectResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	key := req.GetKey()
	dataKey := db.DataKey([]byte(label), key)

	exists, err := api.db.Exists(dataKey)
	if err != nil {
		return nil, err
	}

	return &pb.ExistObjectResponse{
		Exists: exists,
	}, nil
}

// ExistObjectStream implements ObjectManagerServer.ExistObjectStream
func (api *ObjectAPI) ExistObjectStream(req *pb.ExistObjectStreamRequest, stream pb.ObjectManager_ExistObjectStreamServer) error {
	label, err := extractStringFromContext(stream.Context(), serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return unauthenticatedError(err)
	}

	keys := req.GetKeys()
	keyLength := len(keys)
	if keyLength == 0 {
		return errors.New("no keys given")
	}

	// if only 1 key is requested,
	// we can simply check if that object exists,
	// no need to do it asynchronously in that case
	if keyLength == 1 {
		key := keys[0]
		resp := pb.ExistObjectStreamResponse{Key: key}
		resp.Exists, err = api.db.Exists(db.DataKey([]byte(label), key))
		if err != nil {
			return err
		}
		return stream.SendMsg(&resp)
	}

	jobCount := api.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	inputCh := make(chan []byte, jobCount)
	outputCh := make(chan pb.ExistObjectStreamResponse, jobCount)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// create an errgroup for all the worker routines,
	// including the input one
	group, ctx := errgroup.WithContext(ctx)

	// start the input goroutine,
	// which simply gives all keys over a channel
	group.Go(func() error {
		// close input channel when this goroutine is finished
		// (either because of an error or because all items have been received)
		defer close(inputCh)

		for _, key := range keys {
			select {
			case inputCh <- key:
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	// start the output goroutine,
	// as we are only allowed to send to the stream on a single goroutine
	// (sending on multiple goroutines at once is not safe according to docs)
	group.Go(func() error {
		// local variables reused for each iteration/item
		var (
			resp            pb.ExistObjectStreamResponse
			workerStopCount int
		)

		// loop while we can receive intermediate objects,
		// or until the context is done
		for {
			select {
			case <-ctx.Done():
				return nil // early exist -> context is done
			case resp = <-outputCh:
				if resp.Key == nil {
					workerStopCount++
					if workerStopCount == jobCount {
						return nil // we're done!
					}
					continue
				}
			}
			err := stream.Send(&resp)
			if err != nil {
				return err
			}
		}
	})

	// start all the workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// local variables reused for each iteration/item
			var (
				err  error
				open bool
				resp pb.ExistObjectStreamResponse
			)
			for {
				// get the next object key,
				// or return in case the input channel or context has been closed
				select {
				case resp.Key, open = <-inputCh:
					if !open {
						select {
						// try to return nil response,
						// as to indicate this worker is finished
						case outputCh <- pb.ExistObjectStreamResponse{}:
						case <-ctx.Done():
						}
						return nil
					}
				case <-ctx.Done():
					return nil
				}

				// check if the object exists
				resp.Exists, err = api.db.Exists(db.DataKey([]byte(label), resp.Key))
				if err != nil {
					return err
				}

				// return the response
				select {
				case outputCh <- resp:
				case <-ctx.Done():
					// return, context is done
					return nil
				}
			}
		})
	}

	// wait until all contexts are finished
	return group.Wait()
}

// DeleteObjects implements ObjectManagerServer.DeleteObjects
func (api *ObjectAPI) DeleteObjects(ctx context.Context, req *pb.DeleteObjectsRequest) (*pb.DeleteObjectsResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	keys := req.GetKeys()
	keyLength := len(keys)
	if keyLength == 0 {
		return nil, errors.New("no keys given")
	}

	// if we're only dealing with one key,
	// we can delete it directly
	if keyLength == 1 {
		dataKey := db.DataKey([]byte(label), keys[0])
		err = api.db.Delete(dataKey)
		if err != nil {
			return nil, err
		}
		return &pb.DeleteObjectsResponse{}, nil
	}

	// if we're deleting more then one object,
	// let's do it async

	jobCount := api.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	inputCh := make(chan []byte, jobCount)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// create an errgroup for all the worker routines,
	// including the input one
	group, ctx := errgroup.WithContext(ctx)

	// start the input goroutine,
	// which simply gives all keys over a channel
	group.Go(func() error {
		// close input channel when this goroutine is finished
		// (either because of an error or because all items have been received)
		defer close(inputCh)

		for _, key := range keys {
			select {
			case inputCh <- key:
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	// start all the workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// local variables reused for each iteration/item
			var (
				err  error
				key  []byte
				open bool
			)
			for {
				// get the next object key,
				// or return in case the input channel or context has been closed
				select {
				case key, open = <-inputCh:
					if !open {
						return nil
					}
				case <-ctx.Done():
					return nil
				}

				// Delete Object
				err = api.db.Delete(db.DataKey([]byte(label), key))
				if err != nil {
					return err
				}
			}
		})
	}

	// wait until all contexts are finished
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	// all was deleted successfully
	return &pb.DeleteObjectsResponse{}, nil
}

// GetObjectStatus implements ObjectManagerServer.GetObjectStatus
func (api *ObjectAPI) GetObjectStatus(ctx context.Context, req *pb.GetObjectStatusRequest) (*pb.GetObjectStatusResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	key := req.GetKey()
	status, err := serverAPI.ObjectStatusForObject([]byte(label), key, api.db)
	if err != nil {
		return nil, err
	}
	return &pb.GetObjectStatusResponse{Status: convertStatus(status)}, nil

}

// GetObjectStatusStream implements ObjectManagerServer.GetObjectStatusStream
func (api *ObjectAPI) GetObjectStatusStream(req *pb.GetObjectStatusStreamRequest, stream pb.ObjectManager_GetObjectStatusStreamServer) error {
	label, err := extractStringFromContext(stream.Context(), serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return unauthenticatedError(err)
	}

	keys := req.GetKeys()
	keyLength := len(keys)
	if keyLength == 0 {
		return errors.New("no keys given")
	}

	// if we're only dealing with one key,
	// we can check the status for this single object directly
	if keyLength == 1 {
		resp := pb.GetObjectStatusStreamResponse{Key: keys[0]}
		var status server.ObjectStatus
		status, err = serverAPI.ObjectStatusForObject([]byte(label), resp.Key, api.db)
		if err != nil {
			return err
		}
		resp.Status = convertStatus(status)
		return stream.SendMsg(&resp)
	}

	// we're dealing with multiple objects,
	// let's handle them asynchronously

	jobCount := api.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	inputCh := make(chan []byte, jobCount)
	outputCh := make(chan pb.GetObjectStatusStreamResponse, jobCount)

	// create an errgroup for all the worker routines,
	// including the input one
	group, ctx := errgroup.WithContext(ctx)

	// start the input goroutine,
	// so it can start dispatch keys ASAP
	group.Go(func() error {
		// close worker channel when this channel is closed
		// (either because of an error or because all items have been received)
		defer close(inputCh)
		for _, key := range keys {
			select {
			case inputCh <- key:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	// start the output goroutine,
	// as we are only allowed to send to the stream on a single goroutine
	// (sending on multiple goroutines at once is not safe according to docs)
	group.Go(func() error {
		// local variables reused for each iteration/item
		var (
			resp            pb.GetObjectStatusStreamResponse
			workerStopCount int
		)

		// loop while we can receive intermediate objects,
		// or until the context is done
		for {
			select {
			case <-ctx.Done():
				return nil // early exist -> context is done
			case resp = <-outputCh:
				if resp.Key == nil {
					workerStopCount++
					if workerStopCount == jobCount {
						return nil // we're done!
					}
					continue
				}
			}
			err := stream.Send(&resp)
			if err != nil {
				return err
			}
		}
	})

	// start all the workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// local variables reused for each iteration/item
			var (
				resp   pb.GetObjectStatusStreamResponse
				status server.ObjectStatus
				open   bool
			)

			// loop while we can receive intermediate objects,
			// or until the context is done
			for {
				select {
				case <-ctx.Done():
					return nil // early exist -> context is done
				case resp.Key, open = <-inputCh:
					if !open {
						// send a nil response to indicate a worker is finished
						select {
						case outputCh <- pb.GetObjectStatusStreamResponse{}:
						case <-ctx.Done():
							return nil
						}
						return nil // early exit -> worker channel closed
					}
				}

				// check status
				status, err = serverAPI.ObjectStatusForObject(
					[]byte(label), resp.Key, api.db)
				if err != nil {
					return err
				}
				resp.Status = convertStatus(status)

				// send the response ready for output
				select {
				case outputCh <- resp:
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	// wait until all contexts are finished
	return group.Wait()
}

// ListObjectKeyStream implements ObjectManagerServer.ListObjectKeyStream
func (api *ObjectAPI) ListObjectKeyStream(req *pb.ListObjectKeyStreamRequest, stream pb.ObjectManager_ListObjectKeyStreamServer) error {
	label, err := extractStringFromContext(stream.Context(), serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return unauthenticatedError(err)
	}

	// we're dealing with multiple objects,
	// let's handle them asynchronously

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	ch, err := api.db.ListItems(ctx, db.DataPrefix([]byte(label)))
	if err != nil {
		return err
	}

	prefixLength := db.DataKeyPrefixLength([]byte(label))
	outputCh := make(chan pb.ListObjectKeyStreamResponse, api.jobCount)

	// create an errgroup for all the worker routines,
	// including the input one
	group, ctx := errgroup.WithContext(ctx)

	// start the input goroutine,
	// so it can start fetching keys ASAP
	group.Go(func() error {
		// local variables reused for each iteration/item
		var (
			err  error
			key  []byte
			resp pb.ListObjectKeyStreamResponse
		)
		for item := range ch {
			// copy key to take ownership over it
			key = item.Key()
			if len(key) <= prefixLength {
				return errors.New("invalid item key")
			}
			key = key[prefixLength:]
			resp.Key = make([]byte, len(key))
			copy(resp.Key, key)

			// send object over the channel, if possible
			select {
			case outputCh <- resp:
			case <-ctx.Done():
				return nil
			}

			// close current item
			err = item.Close()
			if err != nil {
				return err
			}
		}

		return nil
	})

	// start the output goroutine,
	// as we are only allowed to send to the stream on a single goroutine
	// (sending on multiple goroutines at once is not safe according to docs)
	group.Go(func() error {
		// local variables reused for each iteration/item
		var (
			resp            pb.ListObjectKeyStreamResponse
			workerStopCount int
		)

		// loop while we can receive responses,
		// or until the context is done
		for {
			select {
			case <-ctx.Done():
				return nil // early exist -> context is done
			case resp = <-outputCh:
				if resp.Key == nil {
					workerStopCount++
					if workerStopCount == api.jobCount {
						return nil // we're done!
					}
					continue
				}
			}
			err := stream.Send(&resp)
			if err != nil {
				return err
			}
		}
	})

	// wait until all contexts are finished
	return group.Wait()
}

// SetReferenceList implements ObjectManagerServer.SetReferenceList
func (api *ObjectAPI) SetReferenceList(ctx context.Context, req *pb.SetReferenceListRequest) (*pb.SetReferenceListResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	refList := req.GetReferenceList()
	if refList == nil {
		// if refList is nil,
		// we'll simply try to delete any existent refList,
		// already stored for that key
		refListKey := db.ReferenceListKey([]byte(label), req.GetKey())
		err = api.db.Delete(refListKey)
		return nil, err
	}
	// refList is given, let's encode and store it

	// encode reference list
	data, err := encoding.EncodeReferenceList(refList)
	if err != nil {
		return nil, err
	}

	// store reference list if possible
	refListKey := db.ReferenceListKey([]byte(label), req.GetKey())
	err = api.db.Set(refListKey, data)
	if err != nil {
		return nil, err
	}

	return &pb.SetReferenceListResponse{}, nil
}

// GetReferenceList implements ObjectManagerServer.GetReferenceList
func (api *ObjectAPI) GetReferenceList(ctx context.Context, req *pb.GetReferenceListRequest) (*pb.GetReferenceListResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	var (
		key  = req.GetKey()
		resp pb.GetReferenceListResponse
	)

	refListKey := db.ReferenceListKey([]byte(label), key)
	refListData, err := api.db.Get(refListKey)
	if err != nil {
		if err == db.ErrNotFound {
			resp.Status = pb.ObjectStatusMissing
			return &resp, nil
		}
		return nil, err
	}
	resp.ReferenceList, err = encoding.DecodeReferenceList(refListData)
	if err != nil {
		resp.Status = pb.ObjectStatusCorrupted
		return &resp, nil
	}

	resp.Status = pb.ObjectStatusOK
	return &resp, nil
}

// AppendReferenceList implements ObjectManagerServer.AppendReferenceList
func (api *ObjectAPI) AppendReferenceList(ctx context.Context, req *pb.AppendReferenceListRequest) (*pb.AppendReferenceListResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	// get current reference list data if possible
	refListKey := db.ReferenceListKey([]byte(label), req.GetKey())

	// define update callback
	cb := func(refListData []byte) ([]byte, error) {
		if refListData == nil {
			// if input of update callback is nil, the data didn't exist yet,
			// in which case we can simply encode the target ref list as it is
			return encoding.EncodeReferenceList(req.GetReferenceList())
		}
		// append new list to current list,
		// without decoding the current list
		return encoding.AppendToEncodedReferenceList(refListData, req.GetReferenceList())
	}

	// loop-update until we have no conflict
	err = api.db.Update(refListKey, cb)
	for err == db.ErrConflict {
		err = api.db.Update(refListKey, cb)
	}
	if err != nil {
		if err == encoding.ErrInvalidChecksum || err == encoding.ErrInvalidData {
			return &pb.AppendReferenceListResponse{
				Status: pb.ObjectStatusCorrupted,
			}, nil
		}
		return nil, err
	}

	return &pb.AppendReferenceListResponse{
		Status: pb.ObjectStatusOK,
	}, nil
}

// DeleteReferenceList implements ObjectManagerServer.DeleteReferenceList
func (api *ObjectAPI) DeleteReferenceList(ctx context.Context, req *pb.DeleteReferenceListRequest) (*pb.DeleteReferenceListResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	// get current reference list data if possible
	refListKey := db.ReferenceListKey([]byte(label), req.GetKey())

	// define update callback
	cb := func(refListData []byte) ([]byte, error) {
		if refListData == nil {
			// if input of update callback is nil, the data didn't exist yet,
			// in which case we can simply return nil, as we don't need to do anything
			return nil, nil
		}
		// remove new list from current list
		return encoding.RemoveFromEncodedReferenceList(refListData, req.GetReferenceList())
	}

	// loop-update until we have no conflict
	err = api.db.Update(refListKey, cb)
	for err == db.ErrConflict {
		err = api.db.Update(refListKey, cb)
	}
	if err != nil {
		return nil, err
	}

	return &pb.DeleteReferenceListResponse{}, nil
}

// getObject is a private utility function,
// which centralizes the logic to get an object's value and reference list,
// and return it as it is.
func (api *ObjectAPI) getObject(label, key []byte) (*objectResult, error) {
	object := objectResult{Key: key, Status: pb.ObjectStatusOK}

	// get data
	dataKey := db.DataKey(label, key)
	rawData, err := api.db.Get(dataKey)
	if err != nil {
		if err == db.ErrNotFound {
			object.Status = pb.ObjectStatusMissing
			return &object, nil
		}
		return nil, err
	}
	// decode and validate data
	dataObject, err := encoding.DecodeObject(rawData)
	if err != nil {
		// invalid value -> corrupted object
		object.Status = pb.ObjectStatusCorrupted
		return &object, nil
	}
	object.Value = dataObject.Data

	// get reference list (if it exists)
	refListKey := db.ReferenceListKey(label, key)
	refListData, err := api.db.Get(refListKey)
	if err != nil {
		if err == db.ErrNotFound {
			// no ref list, we can simply return
			return &object, nil
		}
		return nil, err
	}

	// decode existing reference list
	if refList, err := encoding.DecodeReferenceList(refListData); err != nil {
		// invalid ref list -> corrupted object
		object.Status = pb.ObjectStatusCorrupted
	} else {
		// valid ref list
		object.RefList = refList
	}

	// return object
	return &object, nil
}

type objectResult struct {
	Key, Value []byte
	Status     pb.ObjectStatus
	RefList    []string
}

// convertStatus converts server.ObjectStatus to pb.ObjectStatus
func convertStatus(status server.ObjectStatus) pb.ObjectStatus {
	s, ok := _ProtoObjectStatusMapping[status]
	if !ok {
		panic("unknown ObjectStatus")
	}
	return s
}

var _ProtoObjectStatusMapping = map[server.ObjectStatus]pb.ObjectStatus{
	server.ObjectStatusOK:        pb.ObjectStatusOK,
	server.ObjectStatusMissing:   pb.ObjectStatusMissing,
	server.ObjectStatusCorrupted: pb.ObjectStatusCorrupted,
}
