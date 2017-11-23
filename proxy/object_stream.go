package proxy

import (
	//"io"

	pb "github.com/zero-os/0-stor/proxy/pb"
)

// writeStreamReader is io.Reader implementation
// that is needed by WriteStream API
type writeStreamReader struct {
	stream pb.ObjectService_WriteStreamServer
	buff   []byte
}

func newWriteStreamReader(stream pb.ObjectService_WriteStreamServer) (*writeStreamReader, *pb.WriteStreamRequest, error) {
	req, err := stream.Recv()
	if err != nil {
		return nil, nil, err
	}
	return &writeStreamReader{
		stream: stream,
		buff:   req.Value,
	}, req, nil
}

// Read implements io.Reader.Read interface
func (sr *writeStreamReader) Read(dest []byte) (int, error) {
	wantLen := len(dest)
	if len(sr.buff) >= wantLen {
		return sr.getValFromBuf(dest, wantLen)
	}
	for len(sr.buff) < wantLen {
		req, err := sr.stream.Recv()
		if err != nil {
			return 0, err
		}
		sr.buff = append(sr.buff, req.Value...)
	}
	return sr.getValFromBuf(dest, wantLen)
}

// read value from our internal buffer
func (sr *writeStreamReader) getValFromBuf(dest []byte, wantLen int) (int, error) {
	readLen := wantLen
	if len(sr.buff) < wantLen {
		readLen = len(sr.buff)
	}

	copy(dest, sr.buff[:readLen])
	sr.buff = sr.buff[readLen:]
	return readLen, nil
}

// readStreamWriter is io.Writer implementations
// that is needed by ReadStream API
type readStreamWriter struct {
	stream pb.ObjectService_ReadStreamServer
}

func newReadStreamWriter(stream pb.ObjectService_ReadStreamServer) *readStreamWriter {
	return &readStreamWriter{
		stream: stream,
	}
}

func (rsw *readStreamWriter) Write(p []byte) (int, error) {
	err := rsw.stream.Send(&pb.ReadReply{
		Value: p,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
