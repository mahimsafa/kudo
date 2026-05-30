package state

import (
	"encoding/json"
	"fmt"
)

func MarshalCommand(op OpType, data interface{}) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal command data: %w", err)
	}
	cmd := Command{Op: op, Data: raw}
	out, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}
	return out, nil
}
