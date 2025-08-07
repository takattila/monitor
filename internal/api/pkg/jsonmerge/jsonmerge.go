package jsonmerge

import (
	"encoding/json"
)

// MarshalFunc is a function prototype that can be overridden in tests.
type MarshalFunc func(v interface{}) ([]byte, error)

var (
	jsonUnmarshal = json.Unmarshal
	jsonMarshal   = json.Marshal
)

// MergeJSON merges two JSON objects (source into destination).
// A variadic `marshalFunc` is accepted to allow for mocking in tests.
func MergeJSON(source, destination json.RawMessage) (json.RawMessage, error) {
	var srcMap, dstMap map[string]interface{}

	if err := jsonUnmarshal(destination, &dstMap); err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(source, &srcMap); err != nil {
		return nil, err
	}

	for k, v := range srcMap {
		dstMap[k] = v
	}

	merged, err := jsonMarshal(dstMap)
	if err != nil {
		return nil, err
	}

	return merged, nil
}
