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
	"testing"

	"github.com/zero-os/0-stor/client/metastor"
	"github.com/zero-os/0-stor/daemon/api/grpc/rpctypes"
	pb "github.com/zero-os/0-stor/daemon/api/grpc/schema"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestSetMetadata(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.SetMetadata(context.Background(), &pb.SetMetadataRequest{
		Metadata: &pb.Metadata{},
	})
	require.NoError(t, err)
}

func TestSetMetadataError(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.SetMetadata(context.Background(), nil)
	require.Equal(t, rpctypes.ErrGRPCNilMetadata, err)
}

func TestGetMetadata(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.GetMetadata(context.Background(), &pb.GetMetadataRequest{
		Key: []byte("foo"),
	})
	require.NoError(t, err)
}

func TestGetMetadataError(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.GetMetadata(context.Background(), nil)
	require.Equal(t, rpctypes.ErrGRPCNilKey, err)
}

func TestDeleteMetadata(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.DeleteMetadata(context.Background(), &pb.DeleteMetadataRequest{
		Key: []byte("foo"),
	})
	require.NoError(t, err)
}

func TestDeleteMetadataError(t *testing.T) {
	service := newMetadataService(metadataClientStub{})

	_, err := service.DeleteMetadata(context.Background(), nil)
	require.Equal(t, rpctypes.ErrGRPCNilKey, err)
}

type metadataClientStub struct{}

func (stub metadataClientStub) SetMetadata(metadata metastor.Metadata) error {
	return nil
}
func (stub metadataClientStub) GetMetadata(key []byte) (*metastor.Metadata, error) {
	return &metastor.Metadata{}, nil
}
func (stub metadataClientStub) DeleteMetadata(key []byte) error {
	return nil
}

var (
	_ metadataClient = metadataClientStub{}
)
