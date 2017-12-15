package processing

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
)

func TestNopProcessor_ReadWrite(t *testing.T) {
	testProcessorReadWrite(t, NopProcessor{})
}

func TestNopProcessor_ReadWrite_MultiLayer(t *testing.T) {
	testProcessorReadWriteMultiLayer(t, NopProcessor{})
}

func TestNopProcessor_ReadWrite_Async(t *testing.T) {
	testProcessorReadWriteAsync(t, func() Processor {
		return NopProcessor{}
	})
}

// fooPrefixProcessor is a simple processor,
// which simply prefixes input data with 'foo_' when writing,
// and it removes that prefix again when reading
// (after validating that the read input data has that prefix indeed).
//
// The purpose of this processor is to validate our
// processor tests' logic, especially for the more complex
// tests this is useful, as to not waste valuable debugging time,
// while in fact our processor test itself is buggy.
type fooPrefixProcessor struct{}

func (fpp fooPrefixProcessor) WriteProcess(data []byte) ([]byte, error) {
	return append([]byte("foo_"), data...), nil
}

func (fpp fooPrefixProcessor) ReadProcess(data []byte) ([]byte, error) {
	if !bytes.HasPrefix(data, []byte("foo_")) {
		return nil, fmt.Errorf("'%s' does not have prefix 'foo_'", data)
	}
	return bytes.TrimPrefix(data, []byte("foo_")), nil
}

func (fpp fooPrefixProcessor) SharedWriteBuffer() bool { return false }

func (fpp fooPrefixProcessor) SharedReadBuffer() bool { return false }

func TestFooPrefixProcessor_ReadWrite(t *testing.T) {
	testProcessorReadWrite(t, fooPrefixProcessor{})
}

func TestFooPrefixProcessor_ReadWrite_MultiLayer(t *testing.T) {
	testProcessorReadWriteMultiLayer(t, fooPrefixProcessor{})
}

func TestFooPrefixProcessor_ReadWrite_Async(t *testing.T) {
	testProcessorReadWriteAsync(t, func() Processor {
		return fooPrefixProcessor{}
	})
}

func TestNewProcessorChain(t *testing.T) {
	require := require.New(t)

	require.Panics(func() {
		NewProcessorChain(nil)
	}, "no processors given")
	require.Panics(func() {
		NewProcessorChain([]Processor{})
	}, "no processors given")
	require.Panics(func() {
		NewProcessorChain([]Processor{NopProcessor{}})
	}, "< 2 processors given")

	chain := NewProcessorChain([]Processor{NopProcessor{}, NopProcessor{}})
	require.NotNil(chain)

	require.False(chain.SharedReadBuffer(), "nop processor has no shared read buffer")
	require.False(chain.SharedWriteBuffer(), "nop processor has no shared write buffer")

	data, err := chain.WriteProcess([]byte("data"))
	require.NoError(err)
	require.Equal([]byte("data"), data)
	data, err = chain.ReadProcess([]byte("data"))
	require.NoError(err)
	require.Equal([]byte("data"), data)

	data, err = chain.WriteProcess(nil)
	require.NoError(err)
	require.Empty(data)
	data, err = chain.ReadProcess(nil)
	require.NoError(err)
	require.Empty(data)
}

func testProcessorReadWrite(t *testing.T, processor Processor) {
	t.Run("fixed-test-cases", func(t *testing.T) {
		testCases := []string{
			"a",
			"foo",
			"Hello, World!",
			"大家好",
			"This... is my finger :)",
		}
		for _, testCase := range testCases {
			testProcessorReadWriteCycle(t, processor, []byte(testCase))
		}
	})

	t.Run("random-test-cases", func(t *testing.T) {
		for i := 0; i < 64; i++ {
			inputData := make([]byte, mathRand.Int31n(256)+1)
			rand.Read(inputData)
			testProcessorReadWriteCycle(t, processor, inputData)
		}
	})
}

func testProcessorReadWriteCycle(t *testing.T, processor Processor, inputData []byte) {
	require := require.New(t)

	data, err := processor.WriteProcess(inputData)
	require.NoError(err)
	require.NotEmpty(data)

	outputData, err := processor.ReadProcess(data)
	require.NoError(err)
	require.Equal(inputData, outputData)
}

func testProcessorReadWriteMultiLayer(t *testing.T, processor Processor) {
	t.Run("fixed-test-cases", func(t *testing.T) {
		testCases := []string{
			"a",
			"foo",
			"Hello, World!",
			"大家好",
			"This... is my finger :)",
		}
		for _, testCase := range testCases {
			testProcessorReadWriteMultiLayerCycle(t, processor, []byte(testCase))
		}
	})

	t.Run("random-test-cases", func(t *testing.T) {
		for i := 0; i < 64; i++ {
			inputData := make([]byte, mathRand.Int31n(256)+1)
			rand.Read(inputData)
			testProcessorReadWriteMultiLayerCycle(t, processor, inputData)
		}
	})
}

func testProcessorReadWriteMultiLayerCycle(t *testing.T, processor Processor, inputData []byte) {
	for n := 2; n <= 8; n++ {
		t.Run(fmt.Sprintf("%d_times", n), func(t *testing.T) {
			require := require.New(t)

			var (
				err  error
				data = inputData
			)

			// write `n` times
			for i := 0; i < n; i++ {
				data, err = processor.WriteProcess(data)
				require.NoError(err)
				require.NotEmpty(data)

				// ensure to copy our data
				// in case the buffer is shared,
				// otherwise we're going to get weird results
				if processor.SharedWriteBuffer() {
					d := make([]byte, len(data))
					copy(d, data)
					data = d
				}
			}

			// read `n` times
			for i := 0; i < n; i++ {
				data, err = processor.ReadProcess(data)
				require.NoError(err)
				require.NotEmpty(data)

				// ensure to copy our data
				// in case the buffer is shared,
				// otherwise we're going to get weird results
				if processor.SharedReadBuffer() {
					d := make([]byte, len(data))
					copy(d, data)
					data = d
				}
			}

			// ensure the inputData equals the last-read data
			require.Equal(inputData, data)
		})
	}
}

func testProcessorReadWriteAsync(t *testing.T, pc func() Processor) {
	t.Run("fixed-test-cases", func(t *testing.T) {
		testProcessorReadWriteAsyncCycle(t, pc, [][]byte{
			[]byte("a"),
			[]byte("foo"),
			[]byte("Hello, World!"),
			[]byte("大家好"),
			[]byte("This... is my finger :)"),
		})
	})

	t.Run("random-test-cases", func(t *testing.T) {
		var testCases [][]byte
		for i := 0; i < 64; i++ {
			testCase := make([]byte, mathRand.Int31n(256)+1)
			rand.Read(testCase)
			testCases = append(testCases, testCase)
		}
		testProcessorReadWriteAsyncCycle(t, pc, testCases)
	})
}

func testProcessorReadWriteAsyncCycle(t *testing.T, pc func() Processor, inputDataSlice [][]byte) {
	group, ctx := errgroup.WithContext(context.Background())

	inputLength := len(inputDataSlice)

	type (
		inputPackage struct {
			inputIndex int
		}

		processedPackage struct {
			inputIndex int
			dataStr    string
			dataSlice  []byte
		}

		outputPackage struct {
			inputIndex      int
			dataStr         string
			dataSlice       []byte
			outputDataStr   string
			outputDataSlice []byte
		}
	)

	// start the input goroutine, to give all the indices
	inputCh := make(chan inputPackage, inputLength)
	group.Go(func() error {
		defer close(inputCh)
		for index := 0; index < inputLength; index++ {
			select {
			case inputCh <- inputPackage{inputIndex: index}:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	// start a write goroutine, using a write processor
	writeProcessor := pc()
	processedCh := make(chan processedPackage, inputLength)
	group.Go(func() error {
		defer close(processedCh)
		for input := range inputCh {
			inputData := inputDataSlice[input.inputIndex]

			data, err := writeProcessor.WriteProcess(inputData)
			if err != nil {
				return err
			}

			pkg := processedPackage{
				inputIndex: input.inputIndex,
				dataStr:    string(data),
			}

			if writeProcessor.SharedWriteBuffer() {
				pkg.dataSlice = make([]byte, len(data))
				copy(pkg.dataSlice, data)
			} else {
				pkg.dataSlice = data
			}

			select {
			case processedCh <- pkg:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	// start a read goroutine, using a read processor
	readProcessor := pc()
	outputCh := make(chan outputPackage, inputLength)
	group.Go(func() error {
		defer close(outputCh)
		for processed := range processedCh {
			if bytes.Compare([]byte(processed.dataStr), processed.dataSlice) != 0 {
				return fmt.Errorf("index %d: dataStr (%s) and dataSlice (%s) not equal any longer)",
					processed.inputIndex, processed.dataStr, processed.dataSlice)
			}

			data := []byte(processed.dataStr)

			outputData, err := readProcessor.ReadProcess(data)
			if err != nil {
				return err
			}

			inputData := inputDataSlice[processed.inputIndex]
			if bytes.Compare(inputData, outputData) != 0 {
				return fmt.Errorf("index %d: inputData (%s) outputData (%s) not equal)",
					processed.inputIndex, inputData, outputData)
			}

			pkg := outputPackage{
				inputIndex:    processed.inputIndex,
				dataStr:       processed.dataStr,
				dataSlice:     processed.dataSlice,
				outputDataStr: string(outputData),
			}

			if readProcessor.SharedWriteBuffer() {
				pkg.outputDataSlice = make([]byte, len(outputData))
				copy(pkg.outputDataSlice, outputData)
			} else {
				pkg.outputDataSlice = outputData
			}

			select {
			case outputCh <- pkg:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	// start a goroutine, simply to validate the final output of all received input
	group.Go(func() error {
		for output := range outputCh {
			if bytes.Compare([]byte(output.dataStr), output.dataSlice) != 0 {
				return fmt.Errorf("index %d: dataStr (%s) and dataSlice (%s) not equal any longer)",
					output.inputIndex, output.dataStr, output.dataSlice)
			}

			if bytes.Compare([]byte(output.outputDataStr), output.outputDataSlice) != 0 {
				return fmt.Errorf("index %d: outputDataStr (%s) and outputDataSlice (%s) not equal any longer)",
					output.inputIndex, output.outputDataStr, output.outputDataSlice)
			}

			inputData := inputDataSlice[output.inputIndex]
			if bytes.Compare(output.outputDataSlice, inputData) != 0 {
				return fmt.Errorf("index %d: output (%s) and input (%s) not equal any longer)",
					output.inputIndex, output.outputDataSlice, inputData)
			}
		}
		return nil
	})

	err := group.Wait()
	require.NoError(t, err)
}
