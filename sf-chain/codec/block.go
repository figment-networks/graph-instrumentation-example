package codec

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"

	pbcodec "github.com/figment-networks/graph-instrumentation-example/chain/proto"
)

func BlockFromProto(b *pbcodec.Block) (*bstream.Block, error) {
	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	block := &bstream.Block{
		Id:             b.Hash,
		Number:         b.Height,
		PreviousId:     b.PrevHash,
		Timestamp:      time.Unix(0, int64(b.Timestamp)).UTC(),
		LibNum:         b.Height - 1,
		PayloadKind:    pbbstream.Protocol_UNKNOWN, // TODO: Create dummy protocol
		PayloadVersion: 1,
	}

	return bstream.GetBlockPayloadSetter(block, content)
}
