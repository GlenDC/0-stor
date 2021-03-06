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

package cmd

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringsFlag(t *testing.T) {
	require := require.New(t)

	// nil flag should work fine
	var flagA Strings
	require.Empty(flagA.String())
	require.Empty(flagA.Strings())

	// setting our flag should work
	require.NoError(flagA.Set("foo,bar"))
	require.Equal("foo,bar", flagA.String())
	require.Equal([]string{"foo", "bar"}, flagA.Strings())

	// unsetting our flag is possible as well
	require.NoError(flagA.Set(""))
	require.Empty(flagA.String())
	require.Empty(flagA.Strings())

	// no nil-ptr protection for flags
	var nilPtrFlag *Strings
	require.Panics(func() {
		nilPtrFlag.Set("foo")
	}, "there is no nil-pointer protection for flags")
}

func TestListenAddressFlag(t *testing.T) {
	tt := []struct {
		input   string
		err     error
		value   string
		network string
	}{
		{
			// default value
			"",
			nil,
			":8080",
			"tcp",
		},
		{
			"8080",
			nil,
			"8080",
			"unix",
		},
		{
			":8080",
			nil,
			":8080",
			"tcp",
		},
		{
			"127.0.0.1:8080",
			nil,
			"127.0.0.1:8080",
			"tcp",
		},
		{
			"badhost:8080",
			errors.New("host not valid"),
			"",
			"tcp",
		},
		{
			"foo",
			nil,
			"foo",
			"unix",
		},
		{
			"/var/foo",
			nil,
			"/var/foo",
			"unix",
		},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			require := require.New(t)

			var (
				l   ListenAddress
				err error
			)

			if tc.input != "" {
				err = l.Set(tc.input)
			}
			if tc.err != nil {
				require.NotNil(err)
				require.Equal(tc.err.Error(), err.Error())
			} else {
				require.NoError(err)
				require.Equal(tc.network, l.NetworkProtocol())
				require.Equal(tc.value, l.String())
			}
		})
	}
}

func TestProfileModeFlag(t *testing.T) {
	tt := []struct {
		input string
		value string
	}{
		{
			"cpu",
			"cpu",
		},
		{
			"mem",
			"mem",
		},
		{
			"block",
			"block",
		},
		{
			"trace",
			"trace",
		},
		{
			"foo",
			"",
		},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			var p ProfileMode
			err := p.Set(tc.input)
			if tc.value == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.value, p.String())
		})
	}
}

func TestTLSVersionFlag(t *testing.T) {
	tt := []struct {
		input    string
		expected uint16
		value    string
	}{
		{
			"TLS10",
			tls.VersionTLS10,
			"TLS10",
		},
		{
			"tls10",
			tls.VersionTLS10,
			"TLS10",
		},
		{
			"tLS10",
			tls.VersionTLS10,
			"TLS10",
		},
		{
			"TLS11",
			tls.VersionTLS11,
			"TLS11",
		},
		{
			"tls11",
			tls.VersionTLS11,
			"TLS11",
		},
		{
			"TLS12",
			tls.VersionTLS12,
			"TLS12",
		},
		{
			"tls12",
			tls.VersionTLS12,
			"TLS12",
		},
		{
			"foo",
			0,
			"",
		},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			var v TLSVersion
			err := v.Set(tc.input)
			if tc.value == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, v.VersionTLS())
			require.Equal(t, tc.value, v.String())
		})
	}
}
