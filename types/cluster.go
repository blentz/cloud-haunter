package types

import (
	"time"

	dataprocpb "google.golang.org/genproto/googleapis/cloud/dataproc/v1"
)

type ClusterContainer struct {
	Clusters []*Cluster
}

func (c *ClusterContainer) Get(cloudType CloudType) []*Cluster {
	items := []*Cluster{}
	for _, item := range c.Clusters {
		if item.CloudType == cloudType {
			items = append(items, item)
		}
	}
	return items
}

func NewClusterContainer(clusters []*Cluster) *ClusterContainer {
	return &ClusterContainer{clusters}
}

// Cluster represents the Hadoop-based clusters on the cloud providers
type Cluster struct {
	Uuid      string                    `json:"ClusterUuid"`
	Name      string                    `json:"ClusterName"`
	Created   time.Time                 `json:"Created"`
	CloudType CloudType                 `json:"CloudType"`
	Region    string                    `json:"Region"`
	Tags      map[string]string         `json:"Tags"`
	Config    *dataprocpb.ClusterConfig `json:"ClusterConfig"`
	State     State                     `json:"State"`
	Owner     string                    `json:"Owner"`
}

// GetName returns the name of the cluster
func (c Cluster) GetName() string {
	return c.Name
}

// GetOwner returns the owner of the cluster
func (c Cluster) GetOwner() string {
	if val, ok := c.Tags["Owner"]; ok {
		return val
	}
	if val, ok := c.Tags["owner"]; ok {
		return val
	}
	return ""
}

// GetCloudType returns the type of the cloud
func (c Cluster) GetCloudType() CloudType {
	return c.CloudType
}

// GetCreated returns the creation time of the cluster
func (c Cluster) GetCreated() time.Time {
	return c.Created
}

// GetItem returns the cluster struct itself
func (c Cluster) GetItem() interface{} {
	return c
}

// GetType returns the cluster's string representation
func (c Cluster) GetType() string {
	return "cluster"
}

func (c Cluster) GetTags() Tags {
	return c.Tags
}
