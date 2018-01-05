/*
 * Copyright (C) 2017-2018 GIG Technology NV and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package grpc

import (
	"errors"
	"fmt"
	"os"

	"github.com/zero-os/0-stor/client/metastor"
	"github.com/zero-os/0-stor/client/pipeline/storage"
	"github.com/zero-os/0-stor/daemon/api/grpc/rpctypes"
	pb "github.com/zero-os/0-stor/daemon/api/grpc/schema"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// extractStringFromContext extracts a string from grpc context's
func extractStringFromContext(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata found in grpc context")
	}

	slice, ok := md[key]
	if !ok || len(slice) < 1 {
		return "", fmt.Errorf("no found '%s' metadata from grpc context", key)
	}

	return slice[0], nil
}

func convertProtoToInMemoryChunkSlice(chunks []pb.Chunk) []metastor.Chunk {
	n := len(chunks)
	if n == 0 {
		return nil
	}
	imChunks := make([]metastor.Chunk, n)
	for i, c := range chunks {
		chunk := &imChunks[i]
		chunk.Size = c.GetSizeInBytes()
		chunk.Hash = c.GetHash()
		objects := c.GetObjects()
		if n = len(objects); n > 0 {
			chunk.Objects = make([]metastor.Object, n)
			for i, o := range objects {
				object := &chunk.Objects[i]
				object.Key = o.GetKey()
				object.ShardID = o.GetShardID()
			}
		}
	}
	return imChunks
}

func convertProtoToInMemoryMetadata(metadata *pb.Metadata) metastor.Metadata {
	return metastor.Metadata{
		Key:            metadata.GetKey(),
		Size:           metadata.GetSizeInBytes(),
		CreationEpoch:  metadata.GetCreationEpoch(),
		LastWriteEpoch: metadata.GetLastWriteEpoch(),
		Chunks:         convertProtoToInMemoryChunkSlice(metadata.GetChunks()),
	}
}

func convertInMemoryToProtoChunkSlice(chunks []metastor.Chunk) []pb.Chunk {
	n := len(chunks)
	if n == 0 {
		return nil
	}

	protoChunks := make([]pb.Chunk, n)
	for i, c := range chunks {
		chunk := &protoChunks[i]
		chunk.SizeInBytes = c.Size
		chunk.Hash = c.Hash
		if n = len(c.Objects); n > 0 {
			chunk.Objects = make([]pb.Object, n)
			for i, o := range c.Objects {
				object := &chunk.Objects[i]
				object.Key = o.Key
				object.ShardID = o.ShardID
			}
		}
	}
	return protoChunks
}

func convertInMemoryToProtoMetadata(metadata metastor.Metadata) *pb.Metadata {
	return &pb.Metadata{
		Key:            metadata.Key,
		SizeInBytes:    metadata.Size,
		CreationEpoch:  metadata.CreationEpoch,
		LastWriteEpoch: metadata.LastWriteEpoch,
		Chunks:         convertInMemoryToProtoChunkSlice(metadata.Chunks),
	}
}

func convertStorageToProtoCheckStatus(status storage.CheckStatus) pb.CheckStatus {
	checkStatus, ok := _StorageToProtoCheckStatusMapping[status]
	if !ok {
		panic(fmt.Sprintf("unsupported check status: %v", status))
	}
	return checkStatus
}

func openFileToWrite(filePath string, fileMode pb.FileMode, sync bool) (*os.File, error) {
	if len(filePath) == 0 {
		return nil, rpctypes.ErrGRPCNilFilePath
	}

	flags := os.O_RDWR | os.O_CREATE
	switch fileMode {
	case pb.FileModeTruncate:
		flags |= os.O_TRUNC
	case pb.FileModeAppend:
		flags |= os.O_APPEND
	case pb.FileModeExclusive:
		flags |= os.O_EXCL
	default:
		return nil, rpctypes.ErrGRPCInvalidFileMode
	}
	if sync {
		flags |= os.O_SYNC
	}

	return os.OpenFile(filePath, flags, 0644)
}

var _StorageToProtoCheckStatusMapping = map[storage.CheckStatus]pb.CheckStatus{
	storage.CheckStatusInvalid: pb.CheckStatusInvalid,
	storage.CheckStatusOptimal: pb.CheckStatusOptimal,
	storage.CheckStatusValid:   pb.CheckStatusValid,
}
