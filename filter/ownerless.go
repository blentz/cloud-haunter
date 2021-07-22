package operation

import (
	ctx "github.com/blentz/cloud-haunter/context"
	"github.com/blentz/cloud-haunter/types"
	"github.com/blentz/cloud-haunter/utils"
	log "github.com/sirupsen/logrus"
)

func init() {
	ctx.Filters[types.OwnerlessFilter] = ownerless{}
}

type ownerless struct {
}

func (o ownerless) Execute(items []types.CloudItem) []types.CloudItem {
	log.Debugf("[OWNERLESS] Filtering items without tag %s (%d): [%s]", ctx.OwnerLabel, len(items), items)
	return filter("OWNERLESS", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
		switch item.GetItem().(type) {
		case types.Instance:
			inst := item.(*types.Instance)
			if inst.State == types.Terminated {
				log.Debugf("[OWNERLESS] Filter does not apply for cloud item: %s (%s)", item.GetName(), inst.State)
				return true
			}
			match := !utils.IsAnyMatch(inst.Tags, ctx.OwnerLabel)
			log.Debugf("[OWNERLESS] Instance: %s match: %v (%s)", inst.Name, match, inst.State)
			return match
		case types.Stack:
			stack := item.(*types.Stack)
			match := !utils.IsAnyMatch(stack.Tags, ctx.OwnerLabel)
			log.Debugf("[OWNERLESS] Stack: %s match: %v", stack.Name, match)
			return match
		case types.Cluster:
			if item.GetItem().(types.Cluster).State != types.Running {
				log.Debugf("[OWNERLESS] Filter instance, because it's not in RUNNING state: %s", item.GetName())
				return false
			}
			clust := item.(*types.Cluster)
			match := !utils.IsAnyMatch(clust.Tags, ctx.OwnerLabel)
			log.Debugf("[OWNERLESS] Cluster: %s match: %v (%s)", clust.Name, match, clust.State)
			return match
		default:
			log.Fatalf("[OWNERLESS] Filter does not apply for cloud item: %s", item.GetName())
		}
		return true
	})
}
