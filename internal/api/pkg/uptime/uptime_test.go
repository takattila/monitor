package uptime

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiUptimeSuite struct {
		suite.Suite
	}
)

func (a ApiUptimeSuite) TestGetJSON() {
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	JSON := GetJSON()
	a.Contains(JSON, "uptime_info")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)
}

func (a ApiUptimeSuite) TestString() {
	for _, check := range []struct {
		field Up
		want  string
		err   error
	}{
		{
			field: Up{Years: 1},
			want:  "1 year",
			err:   nil,
		},
		{
			field: Up{Years: 2},
			want:  "2 years",
			err:   nil,
		},
		{
			field: Up{Years: 2},
			want:  "2 years",
			err:   nil,
		},
		{
			field: Up{Months: 2},
			want:  "2 months",
			err:   nil,
		},
		{
			field: Up{Weeks: 2},
			want:  "2 weeks",
			err:   nil,
		},
		{
			field: Up{Days: 2},
			want:  "2 days",
			err:   nil,
		},
		{
			field: Up{Hours: 2},
			want:  "2 hours",
			err:   nil,
		},
		{
			field: Up{Minutes: 2},
			want:  "2 minutes",
			err:   nil,
		},
		{
			field: Up{Seconds: 2},
			want:  "2 seconds",
			err:   nil,
		},
		{
			field: Up{Error: fmt.Errorf("%s", "error")},
			want:  "",
			err:   fmt.Errorf("%s", "error"),
		},
	} {
		result, err := check.field.String()
		a.Equal(check.err, err)
		a.Equal(check.want, result)
	}
}

func TestApiUptimeSuite(t *testing.T) {
	suite.Run(t, new(ApiUptimeSuite))
}
