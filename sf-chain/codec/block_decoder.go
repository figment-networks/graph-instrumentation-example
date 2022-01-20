package codec

import (
	"fmt"

	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
)

func blockDecoder(blk *bstream.Block) (interface{}, error) {
	// TODO: Replace the protocol with a correct number
	if blk.Kind() != pbbstream.Protocol_UNKNOWN {
		return nil, fmt.Errorf("expected kind %s, got %s", pbbstream.Protocol_TENDERMINT, blk.Kind())
	}

	if blk.Version() != 1 {
		return nil, fmt.Errorf("this decoder only knows about version 1, got %d", blk.Version())
	}

	// TODO: Replace the code with block decoding
	/*
		block := YOUR_TYPE_OF_BLOCK

		payload, err := blk.Payload.Get()
		if err != nil {
			return nil, fmt.Errorf("unable to get payload from block stream data: %v", err)
		}

		if err := proto.Unmarshal(payload, blockData); err != nil {
			return nil, fmt.Errorf("unable to decode block stream data: %v", err)
		}
	*/

	return nil, nil
}
