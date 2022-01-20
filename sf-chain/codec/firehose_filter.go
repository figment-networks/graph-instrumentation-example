package codec

import (
	"github.com/streamingfast/bstream"
)

type NOOPFilteringPreprocessor struct {
}

func (f *NOOPFilteringPreprocessor) PreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return blk, nil
}
