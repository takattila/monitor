package jsonmerge

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

// A `MergeJSONTestSuite` struktúra, amely magába foglalja a teszteket.
// Be kell ágyaznia a `suite.Suite` típust.
type MergeJSONTestSuite struct {
	suite.Suite
}

// A `TestMergeJSON` egy tesztfüggvény a tesztek futtatásához.
func TestMergeJSON(t *testing.T) {
	suite.Run(t, new(MergeJSONTestSuite))
}

// A `TestMergeJSONSuccess` teszteli a sikeres JSON összefésülését.
func (s *MergeJSONTestSuite) TestMergeJSONSuccess() {
	tests := []struct {
		name        string
		destination string
		source      string
		expected    string
	}{
		{
			name:        "Successful merge with new keys",
			destination: `{"a": 1, "b": "hello"}`,
			source:      `{"c": true, "d": null}`,
			expected:    `{"a": 1, "b": "hello", "c": true, "d": null}`,
		},
		{
			name:        "Successful override with source keys",
			destination: `{"a": 1, "b": "hello"}`,
			source:      `{"b": "world", "c": 3}`,
			expected:    `{"a": 1, "b": "world", "c": 3}`,
		},
		{
			name:        "Multi-level JSON override, destination data preserved",
			destination: `{"user": {"name": "Alice", "id": 123}, "active": true}`,
			source:      `{"user": {"name": "Bob"}}`,
			expected:    `{"user": {"name": "Bob"}, "active": true}`,
		},
		{
			name:        "Empty source JSON",
			destination: `{"a": 1}`,
			source:      `{}`,
			expected:    `{"a": 1}`,
		},
		{
			name:        "Empty destination JSON",
			destination: `{}`,
			source:      `{"a": 1}`,
			expected:    `{"a": 1}`,
		},
		{
			name:        "Successful merge of different data types",
			destination: `{"str": "value", "num": 10}`,
			source:      `{"bool": true, "array": ["one", "two"]}`,
			expected:    `{"str": "value", "num": 10, "bool": true, "array": ["one", "two"]}`,
		},
		{
			name:        "JSON array override",
			destination: `{"items": [1, 2, 3]}`,
			source:      `{"items": ["a", "b"]}`,
			expected:    `{"items": ["a", "b"]}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceBytes := json.RawMessage(tt.source)
			destBytes := json.RawMessage(tt.destination)

			merged, err := MergeJSON(sourceBytes, destBytes)
			s.Require().NoError(err)

			var mergedMap, expectedMap map[string]interface{}
			err = json.Unmarshal(merged, &mergedMap)
			s.Require().NoError(err)
			err = json.Unmarshal([]byte(tt.expected), &expectedMap)
			s.Require().NoError(err)

			s.Assert().Equal(expectedMap, mergedMap)
		})
	}
}

// A `TestMergeJSONErrors` teszteli a hibás bemeneteket.
func (s *MergeJSONTestSuite) TestMergeJSONErrors() {
	tests := []struct {
		name        string
		destination string
		source      string
	}{
		{
			name:        "Invalid destination JSON input",
			destination: `{"a": 1,}`,
			source:      `{"b": 2}`,
		},
		{
			name:        "Invalid source JSON input",
			destination: `{"a": 1}`,
			source:      `{"b": 2,}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceBytes := json.RawMessage(tt.source)
			destBytes := json.RawMessage(tt.destination)

			merged, err := MergeJSON(sourceBytes, destBytes)
			s.Assert().Error(err)
			s.Assert().Nil(merged)
		})
	}
}

// A `TestMergeJSONMarshalError` teszteli a `json.Marshal` hibakezelési ágát.
func (s *MergeJSONTestSuite) TestMergeJSONMarshalError() {
	destination := json.RawMessage(`{"a": 1}`)
	source := json.RawMessage(`{"b": 2}`)

	mockMarshal := func(v interface{}) ([]byte, error) {
		return nil, errors.New("intentional error for testing")
	}

	merged, err := MergeJSON(source, destination, mockMarshal)

	s.Assert().Error(err)
	s.Assert().Nil(merged)
}
