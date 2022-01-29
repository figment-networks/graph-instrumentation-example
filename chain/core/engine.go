package core

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/figment-networks/graph-instrumentation-example/chain/types"
	"github.com/sirupsen/logrus"
)

type Engine struct {
	blockRate time.Duration
	blockChan chan *types.Block
	prevBlock *types.Block
}

func NewEngine(rate int) Engine {
	blockRate := time.Second / time.Duration(rate)

	return Engine{
		blockRate: blockRate,
		blockChan: make(chan *types.Block),
	}
}

func (e *Engine) Initialize(block *types.Block) error {
	e.prevBlock = block
	return nil
}

func (e *Engine) StartBlockProduction(ctx context.Context) {
	logrus.WithField("rate", e.blockRate).Info("starting block producer")

	for {
		select {
		case <-time.Tick(e.blockRate):
			block := e.createBlock()
			e.blockChan <- &block
		case <-ctx.Done():
			logrus.Info("stopping block producer")
			close(e.blockChan)
			return
		}
	}
}

func (e *Engine) Subscription() <-chan *types.Block {
	return e.blockChan
}

func (e *Engine) createBlock() types.Block {
	block := types.Block{
		Timestamp:    time.Now().UTC(),
		Transactions: []types.Transaction{},
	}

	if e.prevBlock != nil {
		block.Height = e.prevBlock.Height + 1
		block.Hash = makeHash(block.Height)
		block.PrevHash = e.prevBlock.Hash
	} else {
		block.Height = 1
		block.Hash = makeHash(block.Height)
		block.PrevHash = makeHash(1)
	}

	for i := uint64(0); i < block.Height%10; i++ {
		tx := types.Transaction{
			Type:     "transfer",
			Hash:     makeHash(fmt.Sprintf("%v-%v", block.Height, i)),
			Sender:   "0xDEADBEEF",
			Receiver: "0xBAAAAAAD",
			Amount:   big.NewInt(int64(i * 1000000000)),
			Fee:      big.NewInt(10000),
			Success:  true,
			Events: []types.Event{
				{
					Type: "token_transfer",
					Attributes: []types.Attribute{
						{Key: "foo", Value: "bar"},
					},
				},
			},
		}

		block.Transactions = append(block.Transactions, tx)
	}

	e.prevBlock = &block
	return block
}
