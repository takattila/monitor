package playground

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type (
	ApiPlaygroundSuite struct {
		suite.Suite
	}
)

func (a ApiPlaygroundSuite) TestPlayground() {
	content := Playground()
	a.NotNil(content)
}

func TestApiPlaygroundSuite(t *testing.T) {
	suite.Run(t, new(ApiPlaygroundSuite))
}
