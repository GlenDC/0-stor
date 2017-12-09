package grpc

import (
	"context"
	"io"
	"math"

	"golang.org/x/sync/errgroup"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/client/data"
	"github.com/zero-os/0-stor/server/api/grpc/rpctypes"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
func NewClient(addr, label, jwtToken string, jobCount int) (*Client, error) {
	if len(addr) == 0 {
		panic("no zstordb address give")
	}
	if len(label) == 0 {
		panic("no label given")
	}

	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		))
	if err != nil {
		return nil, err
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
	}, nil
}

// SetObject implements data.Client.SetObject
func (c *Client) SetObject(object data.Object) error {
	_, err := c.objService.SetObject(c.contextWithMetadata(nil), &pb.SetObjectRequest{
		Key:           object.Key,
		Data:          object.Data,
		ReferenceList: object.ReferenceList,
	})
	return toErr(nil, err)
}

// GetObject implements data.Client.GetObject
func (c *Client) GetObject(key []byte) (*data.Object, error) {
	resp, err := c.objService.GetObject(c.contextWithMetadata(nil),
		&pb.GetObjectRequest{Key: key})
	if err != nil {
		return nil, toErr(nil, err)
	}

	dataObject := &data.Object{
		Key:           key,
		Data:          resp.GetData(),
		ReferenceList: resp.GetReferenceList(),
	}
	if dataObject.Data == nil {
		return nil, data.ErrInvalidResult
	}
	return dataObject, nil
}

// DeleteObject implements data.Client.DeleteObject
func (c *Client) DeleteObject(key []byte) error {
	// delete the objects from the server
	_, err := c.objService.DeleteObject(
		c.contextWithMetadata(nil), &pb.DeleteObjectRequest{Key: key})
	return toErr(nil, err)
}

// GetObjectStatus implements data.Client.GetObjectStatus
func (c *Client) GetObjectStatus(key []byte) (data.ObjectStatus, error) {
	resp, err := c.objService.GetObjectStatus(
		c.contextWithMetadata(nil), &pb.GetObjectStatusRequest{Key: key})
	if err != nil {
		return data.ObjectStatus(0), toErr(nil, err)
	}
	status := convertStatus(resp.GetStatus())
	return status, nil
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
	stream, err := c.objService.ListObjectKeys(ctx,
		&pb.ListObjectKeysRequest{})
	if err != nil {
		return nil, err
	}

	// create output channel and start fetching from the stream
	ch := make(chan data.ObjectKeyResult, c.jobCount)
	for i := 0; i < c.jobCount; i++ {
		group.Go(func() error {
			// fetch all objects possible
			var (
				input *pb.ListObjectKeysResponse
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
					err = toErr(ctx, err)

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
	_, err := c.objService.SetReferenceList(
		c.contextWithMetadata(nil),
		&pb.SetReferenceListRequest{Key: key, ReferenceList: refList})
	return toErr(nil, err)
}

// GetReferenceList implements data.Client.GetReferenceList
func (c *Client) GetReferenceList(key []byte) ([]string, error) {
	resp, err := c.objService.GetReferenceList(
		c.contextWithMetadata(nil), &pb.GetReferenceListRequest{Key: key})
	if err != nil {
		return nil, toErr(nil, err)
	}
	refList := resp.GetReferenceList()
	if refList == nil {
		return nil, data.ErrInvalidResult
	}
	return refList, nil
}

// GetReferenceCount implements data.Client.GetReferenceCount
func (c *Client) GetReferenceCount(key []byte) (int64, error) {
	resp, err := c.objService.GetReferenceCount(
		c.contextWithMetadata(nil), &pb.GetReferenceCountRequest{Key: key})
	if err != nil {
		return 0, toErr(nil, err)
	}
	return resp.GetCount(), nil
}

// AppendToReferenceList implements data.Client.AppendToReferenceList
func (c *Client) AppendToReferenceList(key []byte, refList []string) error {
	_, err := c.objService.AppendToReferenceList(
		c.contextWithMetadata(nil),
		&pb.AppendToReferenceListRequest{Key: key, ReferenceList: refList})
	return toErr(nil, err)
}

// DeleteFromReferenceList implements data.Client.DeleteFromReferenceList
func (c *Client) DeleteFromReferenceList(key []byte, refList []string) (int64, error) {
	resp, err := c.objService.DeleteFromReferenceList(
		c.contextWithMetadata(nil),
		&pb.DeleteFromReferenceListRequest{Key: key, ReferenceList: refList})
	if err != nil {
		return 0, toErr(nil, err)
	}
	return resp.GetCount(), nil
}

// DeleteReferenceList implements data.Client.DeleteReferenceList
func (c *Client) DeleteReferenceList(key []byte) error {
	_, err := c.objService.DeleteReferenceList(
		c.contextWithMetadata(nil), &pb.DeleteReferenceListRequest{Key: key})
	return toErr(nil, err)
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
			rpctypes.MetaAuthKey, c.jwtToken,
			rpctypes.MetaLabelKey, c.label)
	} else {
		md = metadata.Pairs(rpctypes.MetaLabelKey, c.label)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func toErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	err = rpctypes.Error(err)
	if _, ok := err.(rpctypes.ZStorError); ok {
		return err
	}
	if ctx == nil {
		return err
	}
	code := grpc.Code(err)
	switch code {
	case codes.DeadlineExceeded, codes.Canceled:
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	case codes.FailedPrecondition:
		err = grpc.ErrClientConnClosing
	}
	return err
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
