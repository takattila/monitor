package jsonmerge

import (
	"encoding/json"
)

// MergeJSON merges two JSON objects (source into destination).
func MergeJSON(source, destination json.RawMessage) (json.RawMessage, error) {
	var srcMap, dstMap map[string]interface{}

	if err := json.Unmarshal(destination, &dstMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(source, &srcMap); err != nil {
		return nil, err
	}

	for k, v := range srcMap {
		dstMap[k] = v
	}

	merged, err := json.Marshal(dstMap)
	if err != nil {
		return nil, err
	}

	return merged, nil
}
