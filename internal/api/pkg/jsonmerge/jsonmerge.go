package jsonmerge

import (
	"encoding/json"
)

// MarshalFunc is a function prototype that can be overridden in tests.
type MarshalFunc func(v interface{}) ([]byte, error)

// MergeJSON merges two JSON objects (source into destination).
// A variadic `marshalFunc` is accepted to allow for mocking in tests.
func MergeJSON(source, destination json.RawMessage, marshalFunc ...MarshalFunc) (json.RawMessage, error) {
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

	// Use the provided `marshalFunc` if available, otherwise use the standard `json.Marshal`.
	marshal := json.Marshal
	if len(marshalFunc) > 0 {
		marshal = marshalFunc[0]
	}

	merged, err := marshal(dstMap)
	if err != nil {
		return nil, err
	}

	return merged, nil
}
