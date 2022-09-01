package operation

import (
	"context"
	"fmt"
	"os"
	"time"

	ctx "github.com/blentz/cloud-haunter/context"
	"github.com/blentz/cloud-haunter/types"
	log "github.com/sirupsen/logrus"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spenczar/tdigest"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func init() {
	ctx.Filters[types.IdleFilter] = idle{}
}

type op func(float64) float64

type IdleMetric struct {
	metric     string
	percentile float64
	limit      float64
	measured   float64
	operation  op
}

//get percentile
func getPercentile(datapoints []float64, percentile float64) float64 {
	cnt := 0
	td := tdigest.NewWithCompression(100)
	for _, data := range datapoints {
		cnt++
		td.Add(data, 1)
	}
	return td.Quantile(percentile)
}

//return slice of monitored stats for a specific metric/instance (one month duration)
func getTimeSeriesValue(projectID, metric, instance string) ([]float64, error) {

	var valueType string
	rstats := make([]float64, 0)
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: fmt.Sprintf("metric.type = \"%s\" AND metric.label.instance_name = \"%s\" ", metric, instance),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: time.Now().Add(-30 * 24 * time.Hour).Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: time.Now().Unix(),
			},
		},
	}
	iter := client.ListTimeSeries(ctx, req)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		valueType = resp.ValueType.String()
		for _, point := range resp.Points {
			if valueType == "INT64" {
				rstats = append(rstats, float64(point.Value.GetInt64Value()))
			} else {
				rstats = append(rstats, point.Value.GetDoubleValue())
			}
			rstats = append(rstats, point.Value.GetDoubleValue())
		}
	}
	return rstats, nil
}

func id(x float64) float64 {
	return x
}


// Define what is an Idle instance here
// CPU utilization is less than 0.15 vCPUs for 95% of VM runtime(30 days).
// should return True if instance is idle.
func isIdleFiltered(projectID, instance string) bool {
	idleMetrics := map[string]IdleMetric{
		"usage": {limit: 0.15, percentile: 0.95, metric: "compute.googleapis.com/instance/cpu/utilization", operation: id},
	}
	for _key, idleMetric := range idleMetrics {
		stats, err := getTimeSeriesValue(projectID, idleMetric.metric, instance)
		if err != nil {
			log.Errorln(err)
		}
		copyMetric := idleMetrics[_key]
		copyMetric.measured = copyMetric.operation(getPercentile(stats, copyMetric.percentile))
		idleMetrics[_key] = copyMetric
		log.Printf("metric = %s ,instance = %s ,percentile = %f, limit = %f, measured = %f filtered = %t\n", _key, instance, idleMetrics[_key].percentile, idleMetrics[_key].limit, idleMetrics[_key].measured, idleMetrics[_key].measured < idleMetrics[_key].limit)
	}
	log.Printf("%f,%f", idleMetrics["usage"].measured, idleMetrics["usage"].limit)
	return idleMetrics["usage"].measured < idleMetrics["usage"].limit
}

type idle struct {
}

func (f idle) Execute(items []types.CloudItem) []types.CloudItem {
	log.Debugf("[idle] Filtering items (%d): [%s]", len(items), items)
	return filter("idle", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
		if item.GetCreated().Before(time.Now().Add(-30 * 24 * time.Hour)) {
			log.Printf("[idle] Skipping item %s because it is older than 30 days", item.GetName())
			return false
		} else {
			filtered := isIdleFiltered(os.Getenv("GOOGLE_PROJECT_ID"), item.GetName())
			return filtered
		}
	})
}
