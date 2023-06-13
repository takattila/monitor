package storage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
)

func GetJSON() string {
	storageLines := strings.Split(common.Cli(config.GetStringSlice(Cfg, "on_runtime.commands.storage")), "\n")

	var result string
	jsonArray := make([]string, 0)

	allTotal := uint64(0)
	allActual := uint64(0)
	allFree := uint64(0)
	allPercent := float64(0)

	if len(storageLines) > 0 {
		if Cfg.Data.GetBool("Storage") {

			for _, line := range storageLines {
				if line != "" {
					line = strings.Join(strings.Fields(line), " ")
					jsonArray = append(jsonArray,
						`"`+getStorageName(line)+`": {
							"total": `+fmt.Sprint(common.DynamicSizeIECSize(getTotal(line)))+`,
							"total_unit": "`+fmt.Sprint(common.DynamicSizeIECUnit(getTotal(line)))+`",
							"actual": `+fmt.Sprint(common.DynamicSizeIECSize(getUsed(line)))+`,
							"actual_unit": "`+fmt.Sprint(common.DynamicSizeIECUnit(getUsed(line)))+`",
							"free": `+fmt.Sprint(common.DynamicSizeIECSize(getAvailable(line)))+`,
							"free_unit": "`+fmt.Sprint(common.DynamicSizeIECUnit(getAvailable(line)))+`",
							"percent": `+getPercent(line)+`
						}
					`)
					allTotal += getTotal(line)
					allActual += getUsed(line)
					allFree += getAvailable(line)
				}
			}

			allPercent = common.GetPercent(allActual, allTotal)
		}
	}

	all := `"/all": {
		"total": ` + fmt.Sprint(common.DynamicSizeIECSize(allTotal)) + `,
		"total_unit": "` + fmt.Sprint(common.DynamicSizeIECUnit(allTotal)) + `",
		"actual": ` + fmt.Sprint(common.DynamicSizeIECSize(allActual)) + `,
		"actual_unit": "` + fmt.Sprint(common.DynamicSizeIECUnit(allActual)) + `",
		"free": ` + fmt.Sprint(common.DynamicSizeIECSize(allFree)) + `,
		"free_unit": "` + fmt.Sprint(common.DynamicSizeIECUnit(allFree)) + `",
		"percent": ` + fmt.Sprint(allPercent) + `
	}
	`

	if len(jsonArray) == 0 {
		result = all
	} else {
		result = all + "," + strings.Join(jsonArray, ",")
	}

	return `{ "storage_info": {` + result + `}}`
}

func getStorageName(s string) string {
	ret := "unknown"
	arr := strings.Split(s, " ")
	if len(arr) > 0 {
		ret = arr[5]
	}
	return ret
}

func getTotal(s string) uint64 {
	sizeInt := uint64(0)
	arr := strings.Split(s, " ")
	if len(arr) > 0 {
		size := arr[1]
		sizeInt, _ = strconv.ParseUint(size, 10, 64)
	}
	return sizeInt
}

func getUsed(s string) uint64 {
	sizeInt := uint64(0)
	arr := strings.Split(s, " ")
	if len(arr) > 0 {
		size := arr[2]
		sizeInt, _ = strconv.ParseUint(size, 10, 64)
	}
	return sizeInt
}

func getAvailable(s string) uint64 {
	sizeInt := uint64(0)
	arr := strings.Split(s, " ")
	if len(arr) > 0 {
		size := arr[3]
		sizeInt, _ = strconv.ParseUint(size, 10, 64)
	}
	return sizeInt
}

func getPercent(s string) string {
	ret := 0
	arr := strings.Split(s, " ")
	if len(arr) > 0 {
		percent := arr[4]
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		ret, _ = strconv.Atoi(re.FindAllString(percent, -1)[0])
	}
	return fmt.Sprint(ret)
}
