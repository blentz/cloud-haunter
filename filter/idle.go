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

type IdleMetric struct {
	metric     string
	percentile float64
	limit      float64
	measured   float64
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

//return slice of monitored stats for a specific metric/instance (one week duration)
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
				Seconds: time.Now().Add(-7 * 24 * time.Hour).Unix(),
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

// Define what is an Idle instance here
func IsIdle(projectID, instance string) bool {
	idleMetrics := map[string]IdleMetric{
		"received_bytes": {limit: 156000, percentile: 0.95, metric: "compute.googleapis.com/instance/network/received_packets_count"},
		"usage":          {limit: 0.03, percentile: 0.97, metric: "compute.googleapis.com/instance/cpu/utilization"},
		"sent_bytes":     {limit: 60000, percentile: 0.95, metric: "compute.googleapis.com/instance/network/sent_bytes_count"},
	}
	for _key, idleMetric := range idleMetrics {
		stats, err := getTimeSeriesValue(projectID, idleMetric.metric, instance)
		if err != nil {
			log.Errorln(err)
		}
		copyMetric := idleMetrics[_key]
		copyMetric.measured = getPercentile(stats, copyMetric.percentile)
		idleMetrics[_key] = copyMetric
		log.Printf("metric : %s ,instance = %s ,percentile = %f, limit = %f, measured = %f \n", _key, instance, idleMetrics[_key].percentile, idleMetrics[_key].limit, idleMetrics[_key].measured)
	}
	return idleMetrics["usage"].measured < idleMetrics["usage"].limit && idleMetrics["received_bytes"].measured < idleMetrics["received_bytes"].limit && idleMetrics["sent_bytes"].measured < idleMetrics["sent_bytes"].limit

}

type idle struct {
}

func (f idle) Execute(items []types.CloudItem) []types.CloudItem {
	log.Debugf("[idle] Filtering items (%d): [%s]", len(items), items)
	return filter("idle", items, types.ExclusiveFilter, func(item types.CloudItem) bool {
		filtered := IsIdle(os.Getenv("GOOGLE_PROJECT_ID"), item.GetName())
		return filtered
	})
}
