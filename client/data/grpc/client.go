package grpc

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"

	"github.com/lunny/log"
	"github.com/zero-os/0-stor/client/data"
	"github.com/zero-os/0-stor/server/api"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ (data.Client) = (*Client)(nil)

// Client defines a data client,
// to connect to a zstordb using the GRPC interface.
type Client struct {
	conn             *grpc.ClientConn
	objService       pb.ObjectManagerClient
	namespaceService pb.NamespaceManagerClient

	jobCount int

	jwtToken        string
	jwtTokenDefined bool
	label           string
}

// NewClient create a new data client,
// which allows you to connect to a zstordb using a GRPC interface.
//
// `jobCount` defines the maximum number of jobs that are allowed to be run,
// in all async (iterator) methods of this client.
// If the amount of jobs required is less than the max amount of jobs,
// specified by the `jobCount` parameter, only the jobs required
// will be run.
func NewClient(conn *grpc.ClientConn, label, jwtToken string, jobCount int) *Client {
	// TODO: take an address instead, shouldn't take a connection already
	if conn == nil {
		panic("no connection given")
	}
	if len(label) == 0 {
		panic("no label given")
	}

	if jobCount <= 0 {
		jobCount = data.DefaultJobCount
	}

	return &Client{
		conn:             conn,
		objService:       pb.NewObjectManagerClient(conn),
		namespaceService: pb.NewNamespaceManagerClient(conn),
		jobCount:         jobCount,
		jwtToken:         jwtToken,
		jwtTokenDefined:  len(jwtToken) != 0,
		label:            label,
	}
}

// SetObject implements data.Client.SetObject
func (c *Client) SetObject(object data.Object) error {
	if object.Key == nil {
		return data.ErrNilKey
	}
	if object.Value == nil {
		return data.ErrNilData
	}

	_, err := c.objService.SetObject(c.contextWithMetadata(nil), &pb.SetObjectRequest{
		Key:           object.Key,
		Value:         object.Value,
		ReferenceList: object.ReferenceList,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetObject implements data.Client.GetObject
func (c *Client) GetObject(key []byte) (*data.Object, error) {
	resp, err := c.objService.GetObject(c.contextWithMetadata(nil),
		&pb.GetObjectRequest{Key: key})
	if err != nil {
		return nil, err
	}

	switch resp.GetStatus() {
	case pb.ObjectStatusOK:
	case pb.ObjectStatusMissing:
		return nil, data.ErrNotFound
	case pb.ObjectStatusCorrupted:
		return nil, data.ErrCorrupted
	default:
		return nil, data.ErrInvalidResult
	}

	dataObject := &data.Object{
		Key:           resp.GetKey(),
		Value:         resp.GetValue(),
		ReferenceList: resp.GetReferenceList(),
	}

	if dataObject.Key == nil || dataObject.Value == nil {
		return nil, data.ErrInvalidResult
	}
	return dataObject, nil
}

// GetObjectIterator implements data.Client.GetObjectIterator
func (c *Client) GetObjectIterator(ctx context.Context, keys [][]byte) (<-chan data.ObjectResult, error) {
	// ensure a context is given
	if ctx == nil {
		panic("no context given")
	}

	// validate the key input
	keyLength := len(keys)
	if keyLength == 0 {
		return nil, data.ErrNilKey
	}
	for _, key := range keys {
		if key == nil {
			return nil, data.ErrNilKey
		}
	}

	// define the amount of jobs required
	jobCount := c.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	group, ctx := errgroup.WithContext(ctx)
	ctx = c.contextWithMetadata(ctx)

	// create stream
	stream, err := c.objService.GetObjectStream(ctx, &pb.GetObjectStreamRequest{Keys: keys})
	if err != nil {
		return nil, err
	}

	// create output channel and start fetching from the stream
	ch := make(chan data.ObjectResult, jobCount)
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// fetch all objects possible
			var (
				input      *pb.GetObjectStreamResponse
				key, value []byte
			)
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// receive the next object, and check error as a first task to do
				input, err = stream.Recv()
				if err != nil {
					if err == io.EOF {
						// stream is done
						return nil
					}

					// an unexpected error has happened, exit with an error
					log.Errorf(
						"an error was received while receiving a streamed object for: %v",
						err)
					select {
					case ch <- data.ObjectResult{Error: err}:
					case <-ctx.Done():
					}
					return err
				}

				// Validate the status, key and value
				switch input.GetStatus() {
				case pb.ObjectStatusOK:
					err = nil
				case pb.ObjectStatusMissing:
					err = data.ErrNotFound
				case pb.ObjectStatusCorrupted:
					err = data.ErrCorrupted
				default:
					err = data.ErrInvalidResult
				}
				if err == nil {
					// validate the object
					key = input.GetKey()
					value = input.GetValue()
					if key == nil || value == nil {
						err = data.ErrInvalidResult
					}
				}

				// create the error/valid object result
				var result data.ObjectResult
				if err == nil {
					result.Object = data.Object{
						Key:           key,
						Value:         value,
						ReferenceList: input.GetReferenceList(),
					}
				} else {
					result.Error = err
				}

				// return the result for the given key
				select {
				case ch <- result:
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	// launch the err group, to cancel the context
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf(
				"GetObjectIterator job group has exited with an error: %v",
				err)
		}
	}()

	return ch, nil
}

// ExistObject implements data.Client.ExistObject
func (c *Client) ExistObject(key []byte) (bool, error) {
	if key == nil {
		return false, data.ErrNilKey
	}

	resp, err := c.objService.ExistObject(
		c.contextWithMetadata(nil), &pb.ExistObjectRequest{Key: key})
	if err != nil {
		return false, err
	}
	return resp.GetExists(), nil
}

// ExistObjectIterator implements data.Client.ExistObjectIterator
func (c *Client) ExistObjectIterator(ctx context.Context, keys [][]byte) (<-chan data.ObjectExistResult, error) {
	// ensure a context is given
	if ctx == nil {
		panic("no context given")
	}

	// validate the key input
	keyLength := len(keys)
	if keyLength == 0 {
		return nil, data.ErrNilKey
	}
	for _, key := range keys {
		if key == nil {
			return nil, data.ErrNilKey
		}
	}

	// define the amount of jobs required
	jobCount := c.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	group, ctx := errgroup.WithContext(ctx)
	ctx = c.contextWithMetadata(ctx)

	// create stream
	stream, err := c.objService.ExistObjectStream(ctx, &pb.ExistObjectStreamRequest{Keys: keys})
	if err != nil {
		return nil, err
	}

	// create output channel and start fetching from the stream
	ch := make(chan data.ObjectExistResult, jobCount)
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// fetch all objects possible
			var (
				input *pb.ExistObjectStreamResponse
			)
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// receive the next object, and check error as a first task to do
				input, err = stream.Recv()
				if err != nil {
					if err == io.EOF {
						// stream is done
						return nil
					}

					// an unexpected error has happened, exit with an error
					log.Errorf(
						"an error was received while receiving the exists state of an object for: %v",
						err)
					select {
					case ch <- data.ObjectExistResult{Error: err}:
					case <-ctx.Done():
					}
					return err
				}

				// create the error/valid exists result
				result := data.ObjectExistResult{Key: input.GetKey()}
				if result.Key == nil {
					result.Error = data.ErrInvalidResult
				} else {
					result.Exists = input.GetExists()
				}

				// return the result for the given key
				select {
				case ch <- result:
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	// launch the err group, to cancel the context
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf(
				"ExistObjectIterator job group has exited with an error: %v",
				err)
		}
	}()

	return ch, nil
}

// DeleteObjects implements data.Client.DeleteObjects
func (c *Client) DeleteObjects(keys ...[]byte) error {
	// validate the key input
	keyLength := len(keys)
	if keyLength == 0 {
		return data.ErrNilKey
	}
	for _, key := range keys {
		if key == nil {
			return data.ErrNilKey
		}
	}

	// delete the objects from the server
	_, err := c.objService.DeleteObjects(
		c.contextWithMetadata(nil), &pb.DeleteObjectsRequest{Keys: keys})
	return err
}

// GetObjectStatus implements data.Client.GetObjectStatus
func (c *Client) GetObjectStatus(key []byte) (data.ObjectStatus, error) {
	if key == nil {
		return data.ObjectStatus(0), data.ErrNilKey
	}

	resp, err := c.objService.GetObjectStatus(
		c.contextWithMetadata(nil), &pb.GetObjectStatusRequest{Key: key})
	status := convertStatus(resp.GetStatus())
	return status, err
}

// GetObjectStatusIterator implements data.Client.GetObjectStatusIterator
func (c *Client) GetObjectStatusIterator(ctx context.Context, keys [][]byte) (<-chan data.ObjectStatusResult, error) {
	// ensure a context is given
	if ctx == nil {
		panic("no context given")
	}

	// validate the key input
	keyLength := len(keys)
	if keyLength == 0 {
		return nil, data.ErrNilKey
	}
	for _, key := range keys {
		if key == nil {
			return nil, data.ErrNilKey
		}
	}

	// define the amount of jobs required
	jobCount := c.jobCount
	if jobCount > keyLength {
		jobCount = keyLength
	}

	group, ctx := errgroup.WithContext(ctx)
	ctx = c.contextWithMetadata(ctx)

	// create stream
	stream, err := c.objService.GetObjectStatusStream(ctx,
		&pb.GetObjectStatusStreamRequest{Keys: keys})
	if err != nil {
		return nil, err
	}

	// create output channel and start fetching from the stream
	ch := make(chan data.ObjectStatusResult, jobCount)
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			// fetch all objects possible
			var (
				input *pb.GetObjectStatusStreamResponse
			)
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// receive the next object, and check error as a first task to do
				input, err = stream.Recv()
				if err != nil {
					if err == io.EOF {
						// stream is done
						return nil
					}

					// an unexpected error has happened, exit with an error
					log.Errorf(
						"an error was received while receiving the exists state of an object for: %v",
						err)
					select {
					case ch <- data.ObjectStatusResult{Error: err}:
					case <-ctx.Done():
					}
					return err
				}

				// create the error/valid data result
				result := data.ObjectStatusResult{Key: input.GetKey()}
				if result.Key == nil {
					result.Error = data.ErrInvalidResult
				} else {
					status := input.GetStatus()
					result.Status = convertStatus(status)
				}

				// return the result for the given key
				select {
				case ch <- result:
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	// launch the err group, to cancel the context
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf(
				"ExistObjectIterator job group has exited with an error: %v",
				err)
		}
	}()

	return ch, nil
}

// GetNamespace implements data.Client.GetNamespace
func (c *Client) GetNamespace() (*data.Namespace, error) {
	resp, err := c.namespaceService.GetNamespace(
		c.contextWithMetadata(nil), &pb.GetNamespaceRequest{})
	if err != nil {
		return nil, err
	}

	ns := &data.Namespace{Label: resp.GetLabel()}
	if len(ns.Label) == 0 {
		return nil, data.ErrInvalidResult
	}

	ns.ReadRequestPerHour = resp.GetReadRequestPerHour()
	ns.WriteRequestPerHour = resp.GetWriteRequestPerHour()
	ns.NrObjects = resp.GetNrObjects()
	return ns, nil
}

// ListObjectKeyIterator implements data.Client.ListObjectKeyIterator
func (c *Client) ListObjectKeyIterator(ctx context.Context) (<-chan data.ObjectKeyResult, error) {
	// ensure a context is given
	if ctx == nil {
		panic("no context given")
	}

	group, ctx := errgroup.WithContext(ctx)
	ctx = c.contextWithMetadata(ctx)

	// create stream
	stream, err := c.objService.ListObjectKeyStream(ctx,
		&pb.ListObjectKeyStreamRequest{})
	if err != nil {
		return nil, err
	}

	// create output channel and start fetching from the stream
	ch := make(chan data.ObjectKeyResult, c.jobCount)
	for i := 0; i < c.jobCount; i++ {
		group.Go(func() error {
			// fetch all objects possible
			var (
				input *pb.ListObjectKeyStreamResponse
			)
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// receive the next object, and check error as a first task to do
				input, err = stream.Recv()
				if err != nil {
					if err == io.EOF {
						// stream is done
						return nil
					}

					// an unexpected error has happened, exit with an error
					log.Errorf(
						"an error was received while receiving the exists state of an object for: %v",
						err)
					select {
					case ch <- data.ObjectKeyResult{Error: err}:
					case <-ctx.Done():
					}
					return err
				}

				// create the error/valid data result
				result := data.ObjectKeyResult{Key: input.GetKey()}
				if result.Key == nil {
					result.Error = data.ErrInvalidResult
				}

				// return the result for the given key
				select {
				case ch <- result:
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	// launch the err group, to cancel the context
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf(
				"ExistObjectIterator job group has exited with an error: %v",
				err)
		}
	}()

	return ch, nil
}

// SetReferenceList implements data.Client.SetReferenceList
func (c *Client) SetReferenceList(key []byte, refList []string) error {
	if key == nil {
		return data.ErrNilKey
	}

	_, err := c.objService.SetReferenceList(
		c.contextWithMetadata(nil),
		&pb.SetReferenceListRequest{Key: key, ReferenceList: refList})
	return err
}

// GetReferenceList implements data.Client.GetReferenceList
func (c *Client) GetReferenceList(key []byte) ([]string, error) {
	if key == nil {
		return nil, data.ErrNilKey
	}

	resp, err := c.objService.GetReferenceList(
		c.contextWithMetadata(nil), &pb.GetReferenceListRequest{Key: key})
	if err != nil {
		return nil, err
	}

	switch resp.GetStatus() {
	case pb.ObjectStatusOK:
		refList := resp.GetReferenceList()
		if refList == nil {
			return nil, data.ErrInvalidResult
		}
		return refList, nil

	case pb.ObjectStatusMissing:
		return nil, data.ErrNotFound

	case pb.ObjectStatusCorrupted:
		return nil, data.ErrCorrupted

	default:
		return nil, data.ErrInvalidResult
	}
}

// AppendReferenceList implements data.Client.AppendReferenceList
func (c *Client) AppendReferenceList(key []byte, refList []string) error {
	if key == nil {
		return data.ErrNilKey
	}
	if refList == nil {
		return data.ErrNilData
	}

	resp, err := c.objService.AppendReferenceList(
		c.contextWithMetadata(nil),
		&pb.AppendReferenceListRequest{Key: key, ReferenceList: refList})
	if err != nil {
		return err
	}

	switch resp.GetStatus() {
	case pb.ObjectStatusOK:
		return nil

	case pb.ObjectStatusCorrupted:
		return data.ErrCorrupted

	default:
		return data.ErrInvalidResult
	}
}

// DeleteReferenceList implements data.Client.DeleteReferenceList
func (c *Client) DeleteReferenceList(key []byte, refList []string) error {
	if key == nil {
		return data.ErrNilKey
	}
	if refList == nil {
		return data.ErrNilData
	}

	_, err := c.objService.DeleteReferenceList(
		c.contextWithMetadata(nil),
		&pb.DeleteReferenceListRequest{Key: key, ReferenceList: refList})
	return err
}

// Close implements data.Client.Close
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) contextWithMetadata(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var md metadata.MD
	if c.jwtTokenDefined {
		md = metadata.Pairs(
			api.GRPCMetaAuthKey, c.jwtToken,
			api.GRPCMetaLabelKey, c.label)
	} else {
		md = metadata.Pairs(api.GRPCMetaLabelKey, c.label)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// convertStatus converts pb.ObjectStatus data.ObjectStatus
func convertStatus(status pb.ObjectStatus) data.ObjectStatus {
	s, ok := _ProtoObjectStatusMapping[status]
	if !ok {
		panic("unknown ObjectStatus")
	}
	return s
}

var _ProtoObjectStatusMapping = map[pb.ObjectStatus]data.ObjectStatus{
	pb.ObjectStatusOK:        data.ObjectStatusOK,
	pb.ObjectStatusMissing:   data.ObjectStatusMissing,
	pb.ObjectStatusCorrupted: data.ObjectStatusCorrupted,
}
