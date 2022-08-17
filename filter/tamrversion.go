package operation

import (
	"encoding/json"
	semver "golang.org/x/mod/semver"
	"os"

	ctx "github.com/blentz/cloud-haunter/context"
	"github.com/blentz/cloud-haunter/types"
	log "github.com/sirupsen/logrus"
)

// / versions patched for Log4Shell vuln.
var PATCHED []string = []string{
	"2019.019.2",
	"2020.004.3",
	"2020.012.1",
	"2020.016.7",
	"2020.024.3",
	"2021.002.4",
	"2021.006.4",
	"2021.020.1",
	"2021.021.0",
	"develop-SNAPSHOT",
}

var MIN_VERSION = semver.Build("v2021.021.0") // minimum valid version for semver checks.

type tamrVersionInputs struct {
	path string
	port string
}

func init() {
	// initialize new FilterType objects
	httpPathEnv := os.Getenv("HTTPURL_PATH")
	if len(httpPathEnv) < 1 {
		log.Warn("[TAMR-VERSION] no path found in HTTPURL_PATH environment variable.")
	}
	httpPortEnv := os.Getenv("HTTPURL_PORT")
	if len(httpPortEnv) < 1 {
		log.Info("[TAMR-VERSION] no port found in HTTPURL_PORT environment variable.")
	}
	log.Infof("[TAMR-VERSION] path set to: %s, port set to: %s", httpPathEnv, httpPortEnv)
	ctx.Filters[types.TamrVersionFilter] = tamrVersionInputs{httpPathEnv, httpPortEnv}
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (f tamrVersionInputs) Execute(items []types.CloudItem) []types.CloudItem {
	log.Debugf("[TAMR-VERSION] Filtering items (%d): [%s]", len(items), items)
	return filter("TAMR-VERSION", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
		switch item.GetItem().(type) {
		case types.Instance:
			if item.GetItem().(types.Instance).State != types.Running {
				log.Debugf("[TAMR-VERSION] Filter instance, because it's not in RUNNING state: %s", item.GetName())
				return false
			}
		default:
			log.Warnf("[TAMR-VERSION] Filter does not apply for cloud item: %s", item.GetName())
			return false
		}
		response := item.GetItem().(types.Instance).GetUrl(f.path, f.port)
		switch response.Code {
		case 200:
			var responseBody types.TamrVersion
			err := json.Unmarshal(response.Json, &responseBody)
			if err != nil {
				log.Debugf("[TAMR-VERSION] Filter %s, because response from %s could not be processed; response: %+v", item.GetType(), item.GetName(), response.Json)
				return false
			}
			log.Infof("[TAMR-VERSION] %s is running Tamr version: %s", item.GetName(), responseBody.Version)
			if responseBody.Version != "" && !contains(PATCHED, responseBody.Version) {
				sVersion := "v" + string(responseBody.Version)

				// compare(a,b) is: -1 if a<b; 0 if a==b; 1 if a>b
				if semver.Compare(MIN_VERSION, sVersion) == 1 {
					log.Debugf("[TAMR-VERSION] %s: %s does not have a valid version, response: %+v", item.GetType(), item.GetName(), responseBody.Version)
					return true
				}
			}
			return false
		case 0:
			log.Warnf("[TAMR-VERSION] Cloud item did not respond to a request. Filter does not apply for cloud item: %s", item.GetName())
			return false
		default:
			log.Debugf("[TAMR-VERSION] %s: %s does not have a valid version, response: %+v", item.GetType(), item.GetName(), response.Json)
			return true
		}
	})
}
