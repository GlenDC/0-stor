package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	metastor "github.com/zero-os/0-stor/client/metastor/badger"
	"github.com/zero-os/0-stor/client/pipeline"
)

func TestConfig_ValidateAndSanitize(t *testing.T) {
	var cfg Config

	err := cfg.validateAndSanitize()
	require.Error(t, err)

	cfg.Pipeline = new(pipeline.SingleObjectPipeline)
	err = cfg.validateAndSanitize()
	require.Error(t, err)

	require.Zero(t, cfg.MaxMsgSize)

	cfg.MetaClient = new(metastor.Client)
	err = cfg.validateAndSanitize()
	require.NoError(t, err)

	require.Equal(t, DefaultMaxSizeMsg, cfg.MaxMsgSize)
}
