package all

import (
	"encoding/json"

	"github.com/sagernet/sing-box/common/badjsonmerge"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
)

func GetJSON() string {
	rawJSONs := []json.RawMessage{
		json.RawMessage(model.GetJSON()),
		json.RawMessage(cpu.GetJSON()),
		json.RawMessage(memory.GetJSON()),
		json.RawMessage(storage.GetJSON()),
		json.RawMessage(processes.GetJSON()),
		json.RawMessage(services.GetJSON()),
		json.RawMessage(network.GetJSON()),
	}

	rawDestination := json.RawMessage("{}")
	var err error

	for _, rawSource := range rawJSONs {
		rawDestination, err = badjsonmerge.MergeJSON(rawSource, rawDestination)
		if err != nil {
			rawDestination = []byte("{}")
		}
	}

	return string(rawDestination)
}
