package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestAsyncDataSplitter(t *testing.T) {
	testCases := []struct {
		Input          string
		ChunkSize      int
		ExpectedOutput []string
	}{
		{"", 1, nil},
		{"", 42, nil},
		{"a", 1, []string{"a"}},
		{"a", 256, []string{"a"}},
		{"ab", 1, []string{"a", "b"}},
		{"ab", 256, []string{"ab"}},
		{"abcd", 2, []string{"ab", "cd"}},
		{"abcde", 2, []string{"ab", "cd", "e"}},
	}
	for _, testCase := range testCases {
		input := []byte(testCase.Input)
		output := make([][]byte, len(testCase.ExpectedOutput))
		for i, str := range testCase.ExpectedOutput {
			output[i] = []byte(str)
		}
		testAsyncDataSplitterCycle(t, input, output, testCase.ChunkSize)
	}
}

func testAsyncDataSplitterCycle(t *testing.T, input []byte, output [][]byte, chunkSize int) {
	r := bytes.NewReader(input)
	group, ctx := errgroup.WithContext(context.Background())

	inputCh, splitter := newAsyncDataSplitter(ctx, r, chunkSize, len(output))
	group.Go(splitter)

	outputLength := len(output)
	out := make([][]byte, outputLength)
	group.Go(func() error {
		for input := range inputCh {
			if input.Index < 0 || input.Index >= outputLength {
				return fmt.Errorf("received invalid input index '%d'", input.Index)
			}
			if len(out[input.Index]) != 0 {
				return fmt.Errorf("received double input index '%d'", input.Index)
			}
			out[input.Index] = input.Data
		}
		return nil
	})
	err := group.Wait()
	require.NoError(t, err)
	require.Equal(t, output, out)
}
