package proxy

import (
	"errors"
	"os"

	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/client/meta"
	pb "github.com/zero-os/0-stor/proxy/pb"
	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrNilKeyMeta   = errors.New("both key and meta are nil")
	ErrNilFilePath  = errors.New("nil file path")
	ErrNilKey       = errors.New("nil key")
	ErrEmptyRefList = errors.New("empty reference list")
)

type objectSrv struct {
	client *client.Client
}

func newObjectSrv(client *client.Client) *objectSrv {
	return &objectSrv{
		client: client,
	}
}

func (osrv *objectSrv) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteReply, error) {
	if len(req.Key) == 0 && req.Meta == nil {
		return nil, ErrNilKeyMeta
	}

	meta, err := osrv.client.WriteWithMeta([]byte(req.Key),
		[]byte(req.Value),
		[]byte(req.PrevKey),
		pbMetaToStorMeta(req.PrevMeta),
		pbMetaToStorMeta(req.Meta),
		req.ReferenceList)

	if err != nil {
		log.Errorf("Write key %v error: %v", req.Key, err)
		return nil, err
	}

	return &pb.WriteReply{
		Meta: storMetaToPbMeta(meta),
	}, nil
}

func (osrv *objectSrv) WriteFile(ctx context.Context, req *pb.WriteFileRequest) (*pb.WriteFileReply, error) {
	// check input
	if len(req.Key) == 0 && req.Meta == nil {
		return nil, ErrNilKeyMeta
	}

	if len(req.FilePath) == 0 {
		return nil, ErrNilFilePath
	}

	// open the given file path
	f, err := os.Open(req.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta, err := osrv.client.WriteFWithMeta([]byte(req.Key),
		f,
		[]byte(req.PrevKey),
		pbMetaToStorMeta(req.PrevMeta),
		pbMetaToStorMeta(req.Meta),
		req.ReferenceList)

	if err != nil {
		log.Errorf("Write key %v error: %v", req.Key, err)
		return nil, err
	}

	return &pb.WriteFileReply{
		Meta: storMetaToPbMeta(meta),
	}, nil

}

func (osrv *objectSrv) WriteStream(stream pb.ObjectService_WriteStreamServer) error {
	// creates the reader
	sr, req, err := newWriteStreamReader(stream)
	if err != nil {
		return err
	}

	var meta *meta.Meta

	if req.Meta != nil {
		meta, err = osrv.client.WriteFWithMeta(req.Key, sr, req.PrevKey,
			pbMetaToStorMeta(req.PrevMeta),
			pbMetaToStorMeta(req.Meta),
			req.ReferenceList)
	} else {
		meta, err = osrv.client.WriteF(req.Key, sr, req.ReferenceList)
	}
	if err != nil {
		return err
	}
	return stream.SendAndClose(&pb.WriteStreamReply{
		Meta: storMetaToPbMeta(meta),
	})
}

func (osrv *objectSrv) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadReply, error) {
	// check input
	if len(req.Key) == 0 && req.Meta == nil {
		return nil, ErrNilKeyMeta
	}

	var (
		value   []byte
		refList []string
		err     error
	)
	if req.Meta != nil {
		value, refList, err = osrv.client.ReadWithMeta(pbMetaToStorMeta(req.Meta))
	} else {
		value, refList, err = osrv.client.Read([]byte(req.Key))
	}

	if err != nil {
		return nil, err
	}

	return &pb.ReadReply{
		Value:         value,
		ReferenceList: refList,
	}, nil
}

func (osrv *objectSrv) ReadFile(ctx context.Context, req *pb.ReadFileRequest) (*pb.ReadFileReply, error) {
	// check input
	if len(req.Key) == 0 && req.Meta == nil {
		return nil, ErrNilKeyMeta
	}

	if len(req.FilePath) == 0 {
		return nil, ErrNilFilePath
	}

	// open the given file path
	f, err := os.Create(req.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var refList []string
	if req.Meta != nil {
		panic("need to implement the needed func")
	} else {
		refList, err = osrv.client.ReadF(req.Key, f)
	}

	if err != nil {
		return nil, err
	}

	return &pb.ReadFileReply{
		ReferenceList: refList,
	}, nil
}

func (osrv *objectSrv) ReadStream(req *pb.ReadRequest, stream pb.ObjectService_ReadStreamServer) error {
	if len(req.Key) == 0 {
		return ErrNilKey
	}
	writer := newReadStreamWriter(stream)

	refList, err := osrv.client.ReadF(req.Key, writer)
	if err != nil {
		return err
	}
	return stream.Send(&pb.ReadReply{
		ReferenceList: refList,
	})
}

func (osrv *objectSrv) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.NullReply, error) {
	// check input
	if len(req.Key) == 0 && req.Meta == nil {
		return nil, ErrNilKeyMeta
	}

	var err error

	if req.Meta != nil {
		err = osrv.client.DeleteWithMeta(pbMetaToStorMeta(req.Meta))
	} else {
		err = osrv.client.Delete(req.Key)
	}
	if err != nil {
		return nil, err
	}
	return &pb.NullReply{}, nil
}

func (osrv *objectSrv) Walk(req *pb.WalkRequest, stream pb.ObjectService_WalkServer) error {
	if len(req.StartKey) == 0 {
		return ErrNilKey
	}

	ch := osrv.client.Walk(req.StartKey, req.FromEpoch, req.ToEpoch)

	for wr := range ch {
		if wr.Error != nil {
			return wr.Error
		}

		// TODO : handle this error
		// who closes the channel 'ch'?
		stream.Send(&pb.WalkReply{
			Key:           wr.Key,
			Meta:          storMetaToPbMeta(wr.Meta),
			Value:         wr.Data,
			ReferenceList: wr.RefList,
		})
	}
	return nil
}

func (osrv *objectSrv) AppendReferenceList(ctx context.Context, req *pb.ReferenceListRequest) (*pb.NullReply, error) {
	err := checkReferenceListReq(req)
	if err != nil {
		return nil, err
	}

	if req.Meta != nil {
		err = osrv.client.AppendReferenceListWithMeta(pbMetaToStorMeta(req.Meta), req.ReferenceList)
	} else {
		err = osrv.client.AppendReferenceList(req.Key, req.ReferenceList)
	}

	if err != nil {
		return nil, err
	}

	return &pb.NullReply{}, nil
}

func (osrv *objectSrv) RemoveReferenceList(ctx context.Context, req *pb.ReferenceListRequest) (*pb.NullReply, error) {
	err := checkReferenceListReq(req)
	if err != nil {
		return nil, err
	}

	log.Infof("ref list = %v", req.ReferenceList)

	if req.Meta != nil {
		err = osrv.client.RemoveReferenceListWithMeta(pbMetaToStorMeta(req.Meta), req.ReferenceList)
	} else {
		err = osrv.client.RemoveReferenceList(req.Key, req.ReferenceList)
	}

	if err != nil {
		return nil, err
	}

	return &pb.NullReply{}, nil
}

func (osrv *objectSrv) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckReply, error) {
	if req.Key == nil {
		return nil, ErrNilKey
	}

	status, err := osrv.client.Check(req.Key)
	if err != nil {
		return nil, err
	}

	return &pb.CheckReply{
		Status: storCheckStatusToPb(status),
	}, nil
}

func (osrv *objectSrv) Repair(ctx context.Context, req *pb.RepairRequest) (*pb.NullReply, error) {
	if req.Key == nil {
		return nil, ErrNilKey
	}

	if err := osrv.client.Repair(req.Key); err != nil {
		return nil, err
	}

	return &pb.NullReply{}, nil
}

func storCheckStatusToPb(cs client.CheckStatus) pb.CheckReply_Status {
	switch cs {
	case client.CheckStatusMissing:
		return pb.CheckReply_missing
	case client.CheckStatusOk:
		return pb.CheckReply_ok
	case client.CheckStatusCorrupted:
		return pb.CheckReply_corrupted
	}
	// TODO : find better way to do this
	panic("TODO : find better way")
}

func checkReferenceListReq(req *pb.ReferenceListRequest) error {
	if req.Key == nil && req.Meta == nil {
		return ErrNilKeyMeta
	}
	if len(req.ReferenceList) == 0 {
		return ErrEmptyRefList
	}
	return nil
}

// convert protobuf chunk to 0-stor native chunk data type
func pbChunksToStorChunks(pbChunks []*pb.Chunk) []*meta.Chunk {
	if len(pbChunks) == 0 {
		return nil
	}

	chunks := make([]*meta.Chunk, 0, len(pbChunks))

	for _, c := range pbChunks {
		chunks = append(chunks, &meta.Chunk{
			Size:   c.Size,
			Key:    []byte(c.Key),
			Shards: c.Shards,
		})
	}

	return chunks
}

// convert from protobuf meta to 0-stor native meta data type
func pbMetaToStorMeta(pbMeta *pb.Meta) *meta.Meta {
	if pbMeta == nil {
		return nil
	}

	return &meta.Meta{
		Epoch:     pbMeta.Epoch,
		Key:       []byte(pbMeta.Key),
		Chunks:    pbChunksToStorChunks(pbMeta.Chunks),
		Previous:  []byte(pbMeta.Previous),
		Next:      []byte(pbMeta.Next),
		ConfigPtr: []byte(pbMeta.ConfigPtr),
	}
}

// convert from 0-stor chunk to protobuf chunk data type
func storChunksToPbChunks(chunks []*meta.Chunk) []*pb.Chunk {
	if len(chunks) == 0 {
		return nil
	}

	pbChunks := make([]*pb.Chunk, 0, len(chunks))

	for _, c := range chunks {
		pbChunks = append(pbChunks, &pb.Chunk{
			Size:   c.Size,
			Key:    c.Key,
			Shards: c.Shards,
		})
	}

	return pbChunks
}

// convert from 0-stor native meta to protobuf meta data type
func storMetaToPbMeta(storMeta *meta.Meta) *pb.Meta {
	if storMeta == nil {
		return nil
	}
	return &pb.Meta{
		Epoch:     storMeta.Epoch,
		Key:       storMeta.Key,
		Chunks:    storChunksToPbChunks(storMeta.Chunks),
		Previous:  storMeta.Previous,
		Next:      storMeta.Next,
		ConfigPtr: storMeta.ConfigPtr,
	}
}
