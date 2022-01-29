package codec

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	pbcodec "github.com/figment-networks/graph-instrumentation-example/chain/proto"
	"google.golang.org/protobuf/proto"
)

const (
	LogPrefix = "DMLOG"

	MsgBegin = "BLOCK_BEGIN"
	MsgBlock = "BLOCK"
	MsgEnd   = "BLOCK_END"
)

type LogReader struct {
	prefix    string
	prefixLen int
	lines     chan string
	done      chan interface{}
	parseCtx  *ParseCtx
}

type LogEntry struct {
	Kind string
	Data interface{}
}

type ParseCtx struct {
	Height uint64
	Block  *pbcodec.Block
}

func NewLogReader(lines chan string, prefix string) (*LogReader, error) {
	if prefix == "" {
		prefix = LogPrefix
	}

	return &LogReader{
		prefix:    prefix,
		prefixLen: len(prefix),
		lines:     lines,
		done:      make(chan interface{}),
	}, nil
}

func (r *LogReader) Read() (interface{}, error) {
	return r.next()
}

func (r *LogReader) Close() {
}

func (r *LogReader) Done() <-chan interface{} {
	return r.done
}

func (r *LogReader) next() (interface{}, error) {
	for line := range r.lines {
		data, err := r.parseLine(line)
		if err != nil {
			return nil, err
		}

		if data != nil {
			return data, nil
		}
	}

	return nil, io.EOF
}

func (r *LogReader) parseLine(line string) (interface{}, error) {
	if !strings.HasPrefix(line, r.prefix) {
		return nil, nil
	}

	tokens := strings.Split(line[r.prefixLen+1:], " ")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid log line format: %s", line)
	}

	switch tokens[0] {
	case MsgBegin:
		return nil, r.processMsgBegin(tokens[1:])
	case MsgEnd:
		return r.processMsgEnd(tokens[1:])
	case MsgBlock:
		return nil, r.processMsgBlock(tokens[1:])
	default:
		return nil, fmt.Errorf("unsupported kind: %v", tokens[0])
	}
}

func (r *LogReader) processMsgBegin(tokens []string) error {
	height, err := strconv.ParseUint(tokens[0], 10, 64)
	if err != nil {
		return err
	}

	if r.parseCtx != nil && height < r.parseCtx.Height+1 {
		return fmt.Errorf("unexpected begin message at height %v", height)
	}

	r.parseCtx = &ParseCtx{Height: height}
	return nil
}

func (r *LogReader) processMsgEnd(tokens []string) (interface{}, error) {
	height, err := strconv.ParseUint(tokens[0], 10, 64)
	if err != nil {
		return nil, err
	}

	if r.parseCtx == nil {
		return nil, fmt.Errorf("unexpected end marker at height %v", height)
	}

	if height != r.parseCtx.Height {
		return nil, fmt.Errorf("invalid end marker at height %v", height)
	}

	return r.parseCtx.Block, nil
}

func (r *LogReader) processMsgBlock(tokens []string) error {
	block := &pbcodec.Block{}
	if _, err := parseFromProto(tokens[0], block); err != nil {
		return err
	}

	r.parseCtx.Block = block
	return nil
}

func parseFromProto(data string, message proto.Message) (proto.Message, error) {
	buf, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return message, proto.Unmarshal(buf, message)
}
