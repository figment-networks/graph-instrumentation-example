package codec

import (
	"fmt"
	"io"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dbin"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"
)

type BlockReader struct {
	src *dbin.Reader
}

func NewBlockReader(reader io.Reader) (out *BlockReader, err error) {
	dbinReader := dbin.NewReader(reader)

	contentType, version, err := dbinReader.ReadHeader()
	if err != nil {
		return nil, fmt.Errorf("unable to read file header: %s", err)
	}

	// TODO: Pick a correct protocol number
	Protocol := pbbstream.Protocol(pbbstream.Protocol_value[contentType])
	if Protocol != pbbstream.Protocol_UNKNOWN && version != 1 {
		return nil, fmt.Errorf("reader only knows about %s block kind at version 1, got %s at version %d", Protocol, contentType, version)
	}

	return &BlockReader{
		src: dbinReader,
	}, nil
}

func (l *BlockReader) Read() (*bstream.Block, error) {
	message, err := l.src.ReadMessage()

	if len(message) > 0 {
		pbBlock := new(pbbstream.Block)
		err = proto.Unmarshal(message, pbBlock)
		if err != nil {
			return nil, fmt.Errorf("unable to read block proto: %s", err)
		}

		blk, err := bstream.NewBlockFromProto(pbBlock)
		if err != nil {
			return nil, err
		}

		return blk, nil
	}

	if err == io.EOF {
		return nil, err
	}

	return nil, fmt.Errorf("failed reading next dbin message: %s", err)
}
