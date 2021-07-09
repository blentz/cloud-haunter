package operation

import (
    "os"

    ctx "github.com/blentz/cloud-haunter/context"
    "github.com/blentz/cloud-haunter/types"
    log "github.com/sirupsen/logrus"
)

type httpUrl struct {
    path string
}


func init() {
    // initialize new FilterType objects
    httpPathEnv := os.Getenv("HTTPURL_PATH")
    if len(httpPathEnv) < 1 {
        log.Fatalf("[HTTPURL] no path found in HTTPURL_PATH environment variable.")
        panic("unable to continue.")
    }
    log.Infof("[HTTPURL] path set to: %s", httpPathEnv)
    ctx.Filters[types.HttpUrlFilter] = httpUrl{httpPathEnv}
}

func (f httpUrl) Execute(items []types.CloudItem) []types.CloudItem {
    log.Debugf("[HTTPURL] Filtering instances (%d): [%s]", len(items), items)
    return filter("HTTPURL", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
        switch item.GetItem().(type) {
        case types.Instance:
            if item.GetItem().(types.Instance).State != types.Running {
                log.Debugf("[HTTPURL] Filter instance, because it's not in RUNNING state: %s", item.GetName())
                return false
            }
        default:
            log.Fatalf("[HTTPURL] Filter does not apply for cloud item: %s", item.GetName())
            return true
        }
        match := item.GetUrl()
        log.Debugf("[HTTPURL] %s: %s match: %v", item.GetType(), item.GetName(), match)
        return match
    })
}
