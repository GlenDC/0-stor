package processing

// Processor todo
// A processor is not thread-safe,
// and should only ever be used in a single goroutine at a time.
type Processor interface {
	WriteProcess(input []byte) (output []byte, err error)
	ReadProcess(input []byte) (output []byte, err error)
}

// NopProcessor implements the Processor interface,
// but does not do any processing whatsoever.
// Instead it returns the data it receives.
type NopProcessor struct{}

// WriteProcess implements Processor.WriteProcess
func (nop NopProcessor) WriteProcess(data []byte) ([]byte, error) { return data, nil }

// ReadProcess implements ReadProcess.WriteProcess
func (nop NopProcessor) ReadProcess(data []byte) ([]byte, error) { return data, nil }

func NewProcessorChain(processors []Processor) *ProcessorChain {
	if len(processors) < 2 {
		panic("ProcessorChain requires at least two underlying processors")
	}
	return &ProcessorChain{processors: processors}
}

type ProcessorChain struct {
	processors []Processor
}

// WriteProcess implements Processor.WriteProcess
//
// Processes the given data in the write direction,
// using all the given processors in the order they're given,
// starting with the first and ending with the last.
//
// If an error happens at any point,
// this method will short circuit and return that error immediately.
func (chain *ProcessorChain) WriteProcess(data []byte) ([]byte, error) {
	var err error
	for _, processor := range chain.processors {
		data, err = processor.WriteProcess(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// ReadProcess implements Processor.ReadProcess
//
// Processes the given data in the read direction,
// using all the given processors in the reverse order they're given,
// starting with the last and ending with the first.
//
// If an error happens at any point,
// this method will short circuit and return that error immediately.
func (chain *ProcessorChain) ReadProcess(data []byte) ([]byte, error) {
	var err error
	for i := len(chain.processors) - 1; i >= 0; i-- {
		data, err = chain.processors[i].ReadProcess(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

var (
	_ Processor = NopProcessor{}
	_ Processor = (*ProcessorChain)(nil)
)
