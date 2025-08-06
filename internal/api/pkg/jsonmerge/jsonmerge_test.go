package jsonmerge

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MergeJSONTestSuite struct {
	suite.Suite
	originalMarshal   func(v interface{}) ([]byte, error)
	originalUnmarshal func(data []byte, v interface{}) error
}

func (s *MergeJSONTestSuite) SetupTest() {
	s.originalMarshal = jsonMarshal
	s.originalUnmarshal = jsonUnmarshal
}

func (s *MergeJSONTestSuite) TearDownTest() {
	jsonMarshal = s.originalMarshal
	jsonUnmarshal = s.originalUnmarshal
}

func (s *MergeJSONTestSuite) TestMergeJSON_Success() {
	source := json.RawMessage(`{"b":2,"c":3}`)
	destination := json.RawMessage(`{"a":1,"b":100}`)

	merged, err := MergeJSON(source, destination)
	s.Require().NoError(err)

	var result map[string]interface{}
	err = json.Unmarshal(merged, &result)
	s.Require().NoError(err)

	expected := map[string]interface{}{
		"a": float64(1),
		"b": float64(2),
		"c": float64(3),
	}
	s.Equal(expected, result)
}

func (s *MergeJSONTestSuite) TestMergeJSON_InvalidDestination() {
	source := json.RawMessage(`{"b":2}`)
	destination := json.RawMessage(`{invalid json}`)

	_, err := MergeJSON(source, destination)
	s.Error(err)
}

func (s *MergeJSONTestSuite) TestMergeJSON_InvalidSource() {
	source := json.RawMessage(`{oops}`)
	destination := json.RawMessage(`{"a":1}`)

	_, err := MergeJSON(source, destination)
	s.Error(err)
}

func (s *MergeJSONTestSuite) TestMergeJSON_EmptySource() {
	source := json.RawMessage(`{}`)
	destination := json.RawMessage(`{"a":1}`)

	merged, err := MergeJSON(source, destination)
	s.Require().NoError(err)

	var result map[string]interface{}
	err = json.Unmarshal(merged, &result)
	s.Require().NoError(err)

	expected := map[string]interface{}{
		"a": float64(1),
	}
	s.Equal(expected, result)
}

func (s *MergeJSONTestSuite) TestMergeJSON_EmptyDestination() {
	source := json.RawMessage(`{"x":42}`)
	destination := json.RawMessage(`{}`)

	merged, err := MergeJSON(source, destination)
	s.Require().NoError(err)

	var result map[string]interface{}
	err = json.Unmarshal(merged, &result)
	s.Require().NoError(err)

	expected := map[string]interface{}{
		"x": float64(42),
	}
	s.Equal(expected, result)
}

func (s *MergeJSONTestSuite) TestMergeJSON_MarshalError() {
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("mock marshal error")
	}

	source := json.RawMessage(`{"x":1}`)
	destination := json.RawMessage(`{"y":2}`)

	_, err := MergeJSON(source, destination)
	s.EqualError(err, "mock marshal error")
}

func (s *MergeJSONTestSuite) TestMergeJSON_UnmarshalErrorOnDestination() {
	jsonUnmarshal = func(data []byte, v interface{}) error {
		return errors.New("mock unmarshal error on destination")
	}

	source := json.RawMessage(`{"x":1}`)
	destination := json.RawMessage(`{"y":2}`)

	_, err := MergeJSON(source, destination)
	s.EqualError(err, "mock unmarshal error on destination")
}

func (s *MergeJSONTestSuite) TestMergeJSON_UnmarshalErrorOnSource() {
	jsonUnmarshal = func(data []byte, v interface{}) error {
		if string(data) == `{"y":2}` {
			return nil // destination ok
		}
		return errors.New("mock unmarshal error on source")
	}

	source := json.RawMessage(`{"x":1}`)
	destination := json.RawMessage(`{"y":2}`)

	_, err := MergeJSON(source, destination)
	s.EqualError(err, "mock unmarshal error on source")
}

func TestMergeJSONTestSuite(t *testing.T) {
	suite.Run(t, new(MergeJSONTestSuite))
}
