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
}

func NewLogReader(lines chan string, prefix string) (*LogReader, error) {
	if prefix == "" {
		prefix = "DMLOG"
	}

	return &LogReader{
		prefix:    prefix,
		prefixLen: len(prefix),
		lines:     lines,
		done:      make(chan interface{}),
	}, nil
}

func (r *LogReader) Read() (interface{}, error) {
	for line := range r.lines {
		data, err := r.processLine(line)
		if err != nil {
			return nil, err
		}
		if data != nil {
			return data, nil
		}
	}

	return nil, io.EOF
}

func (r *LogReader) Close() {
}

func (r *LogReader) Done() <-chan interface{} {
	return r.done
}

func (r *LogReader) parseLine(line string) (*LogEntry, error) {
	if !strings.HasPrefix(line, r.prefix) {
		return nil, nil
	}

	tokens := strings.Split(line[r.prefixLen+1:], " ")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid log line format: %s", line)
	}

	entry := LogEntry{
		Kind: tokens[0],
	}

	switch entry.Kind {
	case "BLOCK_BEGIN", "BLOCK_END":
		val, err := strconv.ParseUint(tokens[1], 10, 64)
		if err != nil {
			return nil, err
		}
		entry.Data = val
	case "BLOCK":
		entry.Data = tokens[1]
	case "TX":
		entry.Data = tokens[1]
	default:
		return nil, fmt.Errorf("unsupported kind: %v", entry.Kind)
	}

	return &entry, nil
}

func (r *LogReader) processLine(line string) (interface{}, error) {
	entry, err := r.parseLine(line)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	switch entry.Kind {
	case "BLOCK_BEGIN":
		height := entry.Data.(uint64)

		if r.parseCtx != nil && height < r.parseCtx.Height+1 {
			return nil, fmt.Errorf("unexpected begin message at height %v", height)
		}

		r.parseCtx = &ParseCtx{Height: height}
	case "BLOCK_END":
		height := entry.Data.(uint64)

		if r.parseCtx == nil {
			return nil, fmt.Errorf("unexpected end marker at height %v", height)
		}
		if height != r.parseCtx.Height {
			return nil, fmt.Errorf("invalid end marker at height %v", height)
		}

		return r.parseCtx, nil
	case "BLOCK":
		block := &pbcodec.Block{}
		return parseFromProto(entry.Data.(string), block)
	case "TX":
		tx := &pbcodec.Transaction{}
		return parseFromProto(entry.Data.(string), tx)
	default:
		return nil, fmt.Errorf("unknown message kind %q", entry.Kind)
	}

	return nil, nil
}

func parseFromProto(data string, message proto.Message) (proto.Message, error) {
	buf, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return message, proto.Unmarshal(buf, message)
}
