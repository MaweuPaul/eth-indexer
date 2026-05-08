package decoder

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

// Standard ERC-20 ABI covering Transfer and Approval events
const erc20ABI = `[
    {
        "anonymous": false,
        "inputs": [
            {"indexed": true,  "name": "from",  "type": "address"},
            {"indexed": true,  "name": "to",    "type": "address"},
            {"indexed": false, "name": "value", "type": "uint256"}
        ],
        "name": "Transfer",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {"indexed": true,  "name": "owner",   "type": "address"},
            {"indexed": true,  "name": "spender",  "type": "address"},
            {"indexed": false, "name": "value",    "type": "uint256"}
        ],
        "name": "Approval",
        "type": "event"
    }
]`

type Decoder struct {
	abi abi.ABI
}

type DecodedEvent struct {
	Name   string
	Fields map[string]interface{}
}

func New() (*Decoder, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	return &Decoder{abi: parsedABI}, nil
}

func (d *Decoder) Decode(log types.Log) (*DecodedEvent, error) {
	if len(log.Topics) == 0 {
		return nil, fmt.Errorf("log has no topics")
	}

	event, err := d.abi.EventByID(log.Topics[0])
	if err != nil {
		return nil, fmt.Errorf("unknown event: %w", err)
	}

	fields := make(map[string]interface{})

	// Decode non-indexed fields from Data
	if len(log.Data) > 0 {
		if err := d.abi.UnpackIntoMap(fields, event.Name, log.Data); err != nil {
			return nil, fmt.Errorf("failed to unpack data: %w", err)
		}
	}

	// Decode indexed fields from Topics
	topicIndex := 1
	for _, input := range event.Inputs {
		if input.Indexed && topicIndex < len(log.Topics) {
			fields[input.Name] = log.Topics[topicIndex].Hex()
			topicIndex++
		}
	}

	// Convert big.Int to string for JSON serialization
	for k, v := range fields {
		if b, ok := v.(*big.Int); ok {
			fields[k] = b.String()
		}
	}

	return &DecodedEvent{Name: event.Name, Fields: fields}, nil
}

func (d *DecodedEvent) ToMap() map[string]interface{} {
	result, _ := json.Marshal(d.Fields)
	var m map[string]interface{}
	json.Unmarshal(result, &m)
	m["event"] = d.Name
	return m
}
