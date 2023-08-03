package all

import (
	"encoding/json"

	"github.com/sagernet/sing-box/common/badjsonmerge"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/logos"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/run"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/skins"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/internal/api/pkg/uptime"
)

// AllJSONs holds a json.RawMessage array.
type AllJSONs struct {
	RawJSONs []json.RawMessage
}

// GetJSON returns a merged JSON of all hardware JSONs.
func (r *AllJSONs) GetJSON() string {
	rawDestination := json.RawMessage("{}")
	var err error

	for _, rawSource := range r.RawJSONs {
		rawDestination, err = badjsonmerge.MergeJSON(rawSource, rawDestination)
		if err != nil {
			rawDestination = []byte("{}")
		}
	}

	return string(rawDestination)
}

// GetRawJSONs populates a json.RawMessage array.
func GetRawJSONs() *AllJSONs {
	RawJSONs := []json.RawMessage{
		json.RawMessage(model.GetJSON()),
		json.RawMessage(cpu.GetJSON()),
		json.RawMessage(memory.GetJSON()),
		json.RawMessage(storage.GetJSON()),
		json.RawMessage(processes.GetJSON()),
		json.RawMessage(services.GetJSON()),
		json.RawMessage(network.GetJSON()),
		json.RawMessage(run.GetJSON()),
		json.RawMessage(logos.GetJSON()),
		json.RawMessage(skins.GetJSON()),
		json.RawMessage(uptime.GetJSON()),
	}
	return &AllJSONs{RawJSONs: RawJSONs}
}
