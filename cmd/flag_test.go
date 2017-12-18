package cmd

import (
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
		input string
		err   error
		value string
	}{
		{
			// default value
			"",
			nil,
			":8080",
		},
		{
			"8080",
			errors.New("address 8080: missing port in address"),
			"",
		},
		{
			":8080",
			nil,
			":8080",
		},
		{
			"127.0.0.1:8080",
			nil,
			"127.0.0.1:8080",
		},
		{
			"badhost:8080",
			errors.New("host not valid"),
			"",
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
				require.Equal(tc.err.Error(), err.Error())
			} else {
				require.Equal(tc.value, l.String())
			}
		})
	}
}
