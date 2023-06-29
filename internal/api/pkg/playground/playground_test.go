package playground

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiPlaygroundSuite struct {
		suite.Suite
	}
)

func (a ApiPlaygroundSuite) TestPlayground() {
	L = logger.New(logger.NoneLevel, logger.ColorOff)
	content := Playground()
	a.NotNil(content)
}

func TestApiPlaygroundSuite(t *testing.T) {
	suite.Run(t, new(ApiPlaygroundSuite))
}
