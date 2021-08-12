package operation

import (
	"reflect"

	ctx "github.com/blentz/cloud-haunter/context"
	"github.com/blentz/cloud-haunter/types"
	"github.com/blentz/cloud-haunter/utils"
	log "github.com/sirupsen/logrus"
)

func filter(filterName string, items []types.CloudItem, filterType types.FilterConfigType, isNeeded func(types.CloudItem) bool) []types.CloudItem {
	var filtered []types.CloudItem
	for _, item := range items {
		var include bool
		if filterType.IsInclusive() {
			include = isFilterMatch(filterName, item, filterType, ctx.FilterConfig)
		} else {
			include = !isFilterMatch(filterName, item, filterType, ctx.FilterConfig)
		}
		if include {
			log.Debugf("[%s] item %s is not filtered, because of filter config", filterName, item.GetName())
		} else {
			log.Debugf("[%s] item %s is filtered, because of filter config", filterName, item.GetName())
		}
		if isNeeded(item) && include {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func isFilterMatch(filterName string, item types.CloudItem, filterType types.FilterConfigType, filterConfig types.IFilterConfig) bool {
	name := item.GetName()
	_, ignoreLabelFound := item.GetTags()[ctx.IgnoreLabel]
	if ignoreLabelFound {
		log.Debugf("[%s] Found ignore label on item: %s, label: %s", filterName, name, ctx.IgnoreLabel)
		if ctx.IgnoreLabelDisabled {
			log.Debugf("[%s] Ignore label usage is disabled, continuing to apply filter on item: %s", filterName, name)
		} else {
			if filterType.IsInclusive() {
				log.Debugf("[%s] inclusive filter applied on item: %s", filterName, name)
				return false
			}
			log.Debugf("[%s] exclusive filter applied on item: %s", filterName, name)
			return true
		}
	}

	if filterConfig == nil {
		return false
	}

	var filterEntityType types.FilterEntityType

	switch item.GetItem().(type) {
	case types.Access:
		if filterType.IsInclusive() {
			filterEntityType = types.IncludeAccess
		} else {
			filterEntityType = types.ExcludeAccess
		accessFilter, _, _ := getFilterConfigs(filterConfig, filterType)
		if accessFilter != nil {
			switch item.GetCloudType() {
			case types.AWS:
				return isNameOrOwnerMatch(filterName, item, accessFilter.Aws.Names, accessFilter.Aws.Owners)
			case types.AZURE:
				return isNameOrOwnerMatch(filterName, item, accessFilter.Azure.Names, accessFilter.Azure.Owners)
			case types.GCP:
				return isNameOrOwnerMatch(filterName, item, accessFilter.Gcp.Names, accessFilter.Gcp.Owners)
			default:
				log.Warnf("[%s] Cloud type not supported: %s", filterName, item.GetCloudType())
			}
		}
	case types.Database:
		database := item.GetItem().(types.Database)
		name := item.GetName()
		ignoreLabelFound := utils.IsAnyMatch(database.Tags, ctx.IgnoreLabel)
		if ignoreLabelFound {
			log.Debugf("[%s] Found ignore label on item: %s, label: %s", filterName, name, ctx.IgnoreLabel)
			if ctx.IgnoreLabelDisabled {
				log.Debugf("[%s] Ignore label usage is disabled, continuing to apply filter on item: %s", filterName, name)
			} else {
				if filterType.IsInclusive() {
					log.Debugf("[%s] inclusive filter applied on item: %s", filterName, name)
					return false
				}
				log.Debugf("[%s] exclusive filter applied on item: %s", filterName, name)
				return true
			}
		}
		filtered, applied := applyFilterConfig(filterConfig, filterType, item, filterName, database.Tags)
		if applied {
			return filtered
		}
	case types.Instance, types.Stack, types.Database, types.Disk, types.Alert:
		if filterType.IsInclusive() {
			filterEntityType = types.IncludeInstance
		} else {
			filterEntityType = types.ExcludeInstance
		}
	default:
		log.Warnf("Filtering is not implemented for type %s", reflect.TypeOf(item))
		return false
	case types.Cluster:
		clust := item.GetItem().(types.Cluster)
		name := item.GetName()
		ignoreLabelFound := utils.IsAnyMatch(clust.Tags, ctx.IgnoreLabel)
		if ignoreLabelFound {
			log.Debugf("[%s] Found ignore label on item: %s, label: %s", filterName, name, ctx.IgnoreLabel)
			if ctx.IgnoreLabelDisabled {
				log.Debugf("[%s] Ignore label usage is disabled, continuing to apply filter on item: %s", filterName, name)
			} else {
				if filterType.IsInclusive() {
					log.Debugf("[%s] inclusive filter applied on item: %s", filterName, name)
					return false
				}
				log.Debugf("[%s] exclusive filter applied on item: %s", filterName, name)
				return true
			}
		}
		filtered, applied := applyFilterConfig(filterConfig, filterType, item, filterName, clust.Tags)
		if applied {
			return filtered
		}
	}

	filtered, applied := false, false

	if names := filterConfig.GetFilterValues(filterEntityType, item.GetCloudType(), types.Name); names != nil {
		log.Debugf("[%s] filtering item %s to names [%s]", filterName, item.GetName(), names)
		filtered, applied = filtered || utils.IsStartsWith(item.GetName(), names...), true
        }
}

func applyFilterConfig(filterConfig *types.FilterConfig, filterType types.FilterConfigType, item types.CloudItem, filterName string, tags types.Tags) (applied, filtered bool) {
	_, instanceFilter, _ := getFilterConfigs(filterConfig, filterType)
	if instanceFilter != nil {
		switch item.GetCloudType() {
		case types.AWS:
			return isMatchWithIgnores(filterName, item, tags,
				instanceFilter.Aws.Names, instanceFilter.Aws.Owners, instanceFilter.Aws.Labels), true
		case types.AZURE:
			return isMatchWithIgnores(filterName, item, tags,
				instanceFilter.Azure.Names, instanceFilter.Azure.Owners, instanceFilter.Azure.Labels), true
		case types.GCP:
			return isMatchWithIgnores(filterName, item, tags,
				instanceFilter.Gcp.Names, instanceFilter.Gcp.Owners, instanceFilter.Gcp.Labels), true
		default:
			log.Warnf("[%s] Cloud type not supported: %s", filterName, item.GetCloudType())
		}
	}

	if owners := filterConfig.GetFilterValues(filterEntityType, item.GetCloudType(), types.Owner); owners != nil {
		log.Debugf("[%s] filtering item %s with exact match '%t' to owners [%s]", filterName, item.GetName(), ctx.ExactMatchOwner, owners)
		var ownerMatch bool
		if ctx.ExactMatchOwner {
			ownerMatch = utils.IsAnyEquals(item.GetOwner(), owners...)
		} else {
			ownerMatch = utils.IsStartsWith(item.GetOwner(), owners...)
		}
		filtered, applied = filtered || ownerMatch, true
	}
}

func getFilterConfigs(filterConfig *types.FilterConfig, filterType types.FilterConfigType) (accessConfig *types.FilterAccessConfig, instanceConfig *types.FilterInstanceConfig, clusterConfig *types.FilterClusterConfig) {
	if filterConfig != nil {
		if filterType.IsInclusive() {
			return filterConfig.IncludeAccess, filterConfig.IncludeInstance, filterConfig.IncludeCluster
		}
		return filterConfig.ExcludeAccess, filterConfig.ExcludeInstance, filterConfig.ExcludeCluster
	}
	return nil, nil, nil
}

	if labels := filterConfig.GetFilterValues(filterEntityType, item.GetCloudType(), types.Label); labels != nil {
		log.Debugf("[%s] filtering item %s to labels [%s]", filterName, item.GetName(), labels)
		filtered, applied = filtered || utils.IsAnyStartsWith(item.GetTags(), labels...), true
	}

	if applied {
		if filtered {
			log.Debugf("[%s] item %s matches filter", filterName, item.GetName())
		} else {
			log.Debugf("[%s] item %s does not match filter", filterName, item.GetName())
		}
		return filtered
	} else {
		log.Debugf("[%s] item %s could not be filtered", filterName, item.GetName())
	}

	return false
}
