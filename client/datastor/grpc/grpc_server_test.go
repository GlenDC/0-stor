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
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/memory"
)

func newServerCluster(count int) (*Cluster, func(), error) {
	if count < 1 {
		return nil, nil, errors.New("invalid GRPC server-client count")
	}
	var (
		cleanupSlice []func()
		addressSlice []string
	)
	for i := 0; i < count; i++ {
		_, addr, cleanup, err := newServerClient()
		if err != nil {
			for _, cleanup := range cleanupSlice {
				cleanup()
			}
			return nil, nil, err
		}
		cleanupSlice = append(cleanupSlice, cleanup)
		addressSlice = append(addressSlice, addr)
	}

	cluster, err := NewInsecureCluster(addressSlice, "myLabel", nil)
	if err != nil {
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
		return nil, nil, err
	}

	cleanup := func() {
		cluster.Close()
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
	}
	return cluster, cleanup, nil
}

func newServer() (string, func(), error) {
	return newSecureServer(nil)
}

func newSecureServer(tlsConfig *tls.Config) (string, func(), error) {
	var (
		err      error
		listener net.Listener
	)
	const (
		network = "tcp"
		address = "localhost:0"
	)

	if tlsConfig == nil {
		listener, err = net.Listen(network, address)
	} else {
		listener, err = tls.Listen(network, address, tlsConfig)
	}
	if err != nil {
		return "", nil, err
	}

	server, err := grpc.New(memory.New(), grpc.ServerConfig{})
	if err != nil {
		return "", nil, err
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	cleanup := func() {
		err := server.Close()
		if err != nil {
			panic(err)
		}
	}
	return listener.Addr().String(), cleanup, nil
}

func newServerClient() (*Client, string, func(), error) {
	addr, cleanup, err := newServer()
	if err != nil {
		return nil, "", nil, err
	}

	client, err := NewInsecureClient(addr, "myLabel", nil)
	if err != nil {
		cleanup()
		return nil, "", nil, err
	}

	clean := func() {
		fmt.Sprintln("clean called")
		client.Close()
		cleanup()
	}

	return client, addr, clean, nil
}

func newSecureServerClient() (*Client, string, func(), error) {
	keyStr, certStr, err := genTestCertificate()
	if err != nil {
		return nil, "", nil, err
	}
	cert, err := tls.X509KeyPair([]byte(certStr), []byte(keyStr))
	if err != nil {
		return nil, "", nil, err
	}

	addr, cleanup, err := newSecureServer(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return nil, "", nil, err
	}

	client, err := NewClient(addr, "myLabel", nil, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		cleanup()
		return nil, "", nil, err
	}

	clean := func() {
		fmt.Sprintln("clean called")
		client.Close()
		cleanup()
	}

	return client, addr, clean, nil
}
