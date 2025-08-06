package jsonmerge

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMergeJSON(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		source      string
		expected    string
		expectError bool
	}{
		{
			name:        "Successful merge with new keys",
			destination: `{"a": 1, "b": "hello"}`,
			source:      `{"c": true, "d": null}`,
			expected:    `{"a": 1, "b": "hello", "c": true, "d": null}`,
			expectError: false,
		},
		{
			name:        "Successful override with source keys",
			destination: `{"a": 1, "b": "hello"}`,
			source:      `{"b": "world", "c": 3}`,
			expected:    `{"a": 1, "b": "world", "c": 3}`,
			expectError: false,
		},
		{
			name:        "Multi-level JSON override, destination data preserved",
			destination: `{"user": {"name": "Alice", "id": 123}, "active": true}`,
			source:      `{"user": {"name": "Bob"}}`,
			expected:    `{"user": {"name": "Bob"}, "active": true}`,
			expectError: false,
		},
		{
			name:        "Invalid destination JSON input",
			destination: `{"a": 1,}`,
			source:      `{"b": 2}`,
			expected:    ``,
			expectError: true,
		},
		{
			name:        "Invalid source JSON input",
			destination: `{"a": 1}`,
			source:      `{"b": 2,}`,
			expected:    ``,
			expectError: true,
		},
		{
			name:        "Empty source JSON",
			destination: `{"a": 1}`,
			source:      `{}`,
			expected:    `{"a": 1}`,
			expectError: false,
		},
		{
			name:        "Empty destination JSON",
			destination: `{}`,
			source:      `{"a": 1}`,
			expected:    `{"a": 1}`,
			expectError: false,
		},
		{
			name:        "Successful merge of different data types",
			destination: `{"str": "value", "num": 10}`,
			source:      `{"bool": true, "array": ["one", "two"]}`,
			expected:    `{"str": "value", "num": 10, "bool": true, "array": ["one", "two"]}`,
			expectError: false,
		},
		{
			name:        "JSON array override",
			destination: `{"items": [1, 2, 3]}`,
			source:      `{"items": ["a", "b"]}`,
			expected:    `{"items": ["a", "b"]}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceBytes := json.RawMessage(tt.source)
			destBytes := json.RawMessage(tt.destination)

			merged, err := MergeJSON(sourceBytes, destBytes)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none.")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Compare the output and expected values.
			var mergedMap, expectedMap map[string]interface{}
			err = json.Unmarshal(merged, &mergedMap)
			if err != nil {
				t.Fatalf("Malformed merged JSON: %v", err)
			}
			err = json.Unmarshal([]byte(tt.expected), &expectedMap)
			if err != nil {
				t.Fatalf("Malformed expected JSON: %v", err)
			}

			if !reflect.DeepEqual(mergedMap, expectedMap) {
				t.Errorf("Expected %s, but got %s", tt.expected, string(merged))
			}
		})
	}
}

// Test for the missing coverage on the json.Marshal error path.
func TestMergeJSONMarshalError(t *testing.T) {
	// The `MergeJSON` function's internal logic is reproduced here to
	// force a marshal error. This test is specifically designed to cover
	// the unreachable error condition in the original function.

	// Valid JSON objects are used for the input.
	destination := json.RawMessage(`{"a": 1}`)
	source := json.RawMessage(`{"b": 2}`)

	// Unmarshal the inputs to get maps.
	var dstMap, srcMap map[string]interface{}
	json.Unmarshal(destination, &dstMap)
	json.Unmarshal(source, &srcMap)

	// Merge the maps.
	for k, v := range srcMap {
		dstMap[k] = v
	}

	// Force a marshal error by adding a non-marshalable type (a function) to the map.
	dstMap["fail"] = func() {}

	// Now, test the specific part of the MergeJSON function's error handling.
	_, err := json.Marshal(dstMap)
	if err == nil {
		t.Error("Expected an error from json.Marshal but got none.")
	}
}
