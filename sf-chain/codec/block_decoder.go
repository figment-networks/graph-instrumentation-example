package codec

import (
	"fmt"

	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"

	pbcodec "github.com/figment-networks/graph-instrumentation-example/chain/proto"
)

func blockDecoder(blk *bstream.Block) (interface{}, error) {
	if blk.Kind() != pbbstream.Protocol_UNKNOWN {
		return nil, fmt.Errorf("expected kind %s, got %s", pbbstream.Protocol_TENDERMINT, blk.Kind())
	}

	if blk.Version() != 1 {
		return nil, fmt.Errorf("this decoder only knows about version 1, got %d", blk.Version())
	}

	payload, err := blk.Payload.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to get payload from block stream data: %v", err)
	}

	block := &pbcodec.Block{}
	if err := proto.Unmarshal(payload, block); err != nil {
		return nil, fmt.Errorf("unable to decode block stream data: %v", err)
	}

	return block, nil
}
