package servers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/internal/api/pkg/handlers"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiServersSuite struct {
		suite.Suite
	}
)

func (a ApiServersSuite) TestServeHTTP() {
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	port, err := freeport.GetFreePort()
	if err != nil {
		a.T().Errorf("[ERROR] freeport.GetFreePort: %s\n", err)
	}
	endpoint := "/playground"

	r := chi.NewRouter()
	r.Get(fmt.Sprintf("%s", endpoint), handlers.Playground)
	go ServeHTTP(port, r)
	time.Sleep(100 * time.Millisecond)

	requestURL := fmt.Sprintf("http://localhost:%d%s", port, endpoint)
	res, err := http.Get(requestURL)
	if err != nil {
		a.T().Errorf("[ERROR] http.Get: %s\n", err)
	}

	a.Equal(200, res.StatusCode)
	a.NotNil(res.Body)
}

func TestApiServersSuite(t *testing.T) {
	suite.Run(t, new(ApiServersSuite))
}
