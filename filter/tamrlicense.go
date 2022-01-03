package operation

import (
	"encoding/json"
	"os"
	"time"

	ctx "github.com/blentz/cloud-haunter/context"
	"github.com/blentz/cloud-haunter/types"
	log "github.com/sirupsen/logrus"
)

const LICENSE_GRACE_DAYS = 5 * 24 * time.Hour

type tamrLicenseInputs struct {
	path string
	port string
}

func init() {
	// initialize new FilterType objects
	httpPathEnv := os.Getenv("HTTPURL_PATH")
	if len(httpPathEnv) < 1 {
		log.Warn("[TAMR-LICENSE] no path found in HTTPURL_PATH environment variable.")
	}
	httpPortEnv := os.Getenv("HTTPURL_PORT")
	if len(httpPortEnv) < 1 {
		log.Info("[TAMR-LICENSE] no port found in HTTPURL_PORT environment variable.")
	}
	log.Infof("[TAMR-LICENSE] path set to: %s, port set to: %s", httpPathEnv, httpPortEnv)
	ctx.Filters[types.TamrLicenseFilter] = tamrLicenseInputs{httpPathEnv, httpPortEnv}
}

func (f tamrLicenseInputs) Execute(items []types.CloudItem) []types.CloudItem {
	log.Debugf("[TAMR-LICENSE] Filtering items (%d): [%s]", len(items), items)
	return filter("TAMR-LICENSE", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
		switch item.GetItem().(type) {
		case types.Instance:
			if item.GetItem().(types.Instance).State != types.Running {
				log.Debugf("[TAMR-LICENSE] Filter instance, because it's not in RUNNING state: %s", item.GetName())
				return false
			}
		default:
			log.Fatalf("[TAMR-LICENSE] Filter does not apply for cloud item: %s", item.GetName())
			return false
		}
		version := item.GetItem().(types.Instance).GetUrl("TAMR-LICENSE", "/api/service/version", f.port)
		if version.Error {
			log.Debugf("[TAMR-LICENSE] Filter %s, because %s does not appear to be running Tamr, response: %+v", item.GetType(), item.GetName(), version)
			return false // instance is not unlicensed; instance is probably not running tamr
		} else {
			log.Debugf("[TAMR-LICENSE] %s %s is running Tamr, response: %+v", item.GetType(), item.GetName(), version.Json)
		}
		response := item.GetItem().(types.Instance).GetUrl("TAMR-LICENSE", f.path, f.port)
		switch response.Code {
		case 999:
			log.Debugf("[TAMR-LICENSE] Filter %s, because %s does not appear to be running Tamr, response: %+v", item.GetType(), item.GetName(), response)
			return false // instance is not unlicensed; instance is probably not running tamr
		case 200:
			var tamrBody types.TamrResponseBody
			json.Unmarshal(response.Body, &tamrBody)
			if tamrBody.License.Healthy &&
				tamrBody.License.Message != "tamr license is not valid" &&
				tamrBody.License.Timestamp.Before(time.Now().Add(LICENSE_GRACE_DAYS)) {
				log.Debugf("[TAMR-LICENSE] %s: %s has a valid license until %s",
					item.GetType(),
					item.GetName(),
					tamrBody.License.Timestamp)
			}
			return false
		default:
			log.Debugf("[TAMR-LICENSE] %s: %s does not have a valid license, response: %+v", item.GetType(), item.GetName(), response.Json)
			return true // instance listens on the port, but doesn't give us a valid response.
		}
	})
}
