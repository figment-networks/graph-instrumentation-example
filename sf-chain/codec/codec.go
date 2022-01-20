package codec

import (
	"errors"
	"io"
	"time"

	"github.com/streamingfast/bstream"
)

func init() {
	bstream.GetBlockReaderFactory = bstream.BlockReaderFactoryFunc(BlockReaderFactory)
	bstream.GetBlockDecoder = bstream.BlockDecoderFunc(blockDecoder)
	bstream.GetBlockWriterFactory = bstream.BlockWriterFactoryFunc(BlockWriterFactory)
	bstream.GetBlockWriterHeaderLen = 10
	bstream.GetBlockPayloadSetter = bstream.MemoryBlockPayloadSetter
	bstream.GetMemoizeMaxAge = 200 * 15 * time.Second
}

func Validate() error {
	if err := bstream.ValidateRegistry(); err != nil {
		return err
	}

	if bstream.GetProtocolGenesisBlock == 0 {
		return errors.New("protocol genesis block height must be set")
	}

	if bstream.GetProtocolFirstStreamableBlock == 0 {
		return errors.New("protocol first streamable block must be set")
	}

	return nil
}

func SetProtocolFirstStreamableBlock(height uint64) {
	bstream.GetProtocolFirstStreamableBlock = height
}

func SetProtocolGenesisBlock(height uint64) {
	bstream.GetProtocolGenesisBlock = height
}

func BlockReaderFactory(reader io.Reader) (bstream.BlockReader, error) {
	return NewBlockReader(reader)
}

func BlockWriterFactory(writer io.Writer) (bstream.BlockWriter, error) {
	return NewBlockWriter(writer)
}
