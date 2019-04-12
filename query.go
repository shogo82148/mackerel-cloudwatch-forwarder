package forwarder

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/sirupsen/logrus"
)

// Query is a query for AWS CloudWatch.
type Query struct {
	Service string        `json:"service,omitempty"`
	Host    string        `json:"host,omitempty"`
	Name    string        `json:"name,omitempty"`
	Metric  []interface{} `json:"metric,omitempty"`
	Stat    string        `json:"stat,omitempty"`
}

// ToMetricDataQuery converts the query to cloudwatch.MetricDataQuery.
func ToMetricDataQuery(query []*Query) ([]cloudwatch.MetricDataQuery, error) {
	// Namespace + MetricName + Maximum 10 Dimensions
	var lastMetric [22]string
	var lastHost, lastService, lastStat string

	ret := make([]cloudwatch.MetricDataQuery, 0, len(query))

	for i, q := range query {
		host := q.Host
		setDefault(&host, &lastHost)
		service := q.Service
		setDefault(&service, &lastService)
		stat := q.Stat
		setDefault(&stat, &lastStat)

		if (host == "") == (service == "") {
			logrus.WithFields(logrus.Fields{
				"index":   i,
				"host":    host,
				"service": service,
			}).Warn("either service name or host id is required but not both, skips")
			continue
		}
		if len(q.Metric) < 2 {
			logrus.WithFields(logrus.Fields{
				"index":  i,
				"metric": q.Metric,
			}).Warn("at least, namespace and metric name are required, skips")
		}
		namespace := interfaceToString(q.Metric[0])
		setDefault(&namespace, &lastMetric[0])
		name := interfaceToString(q.Metric[1])
		setDefault(&name, &lastMetric[1])

		var dimensions []cloudwatch.Dimension
		for j := 2; j+1 < len(q.Metric); j += 2 {
			name := interfaceToString(q.Metric[j])
			setDefault(&name, &lastMetric[j])
			value := interfaceToString(q.Metric[j+1])
			setDefault(&value, &lastMetric[j+1])
			dimensions = append(dimensions, cloudwatch.Dimension{
				Name:  aws.String(name),
				Value: aws.String(value),
			})
		}

		label := Label{
			Service:    service,
			HostID:     host,
			MetricName: q.Name,
		}
		metric := &cloudwatch.Metric{
			Namespace:  aws.String(namespace),
			MetricName: aws.String(name),
			Dimensions: dimensions,
		}
		ret = append(ret, cloudwatch.MetricDataQuery{
			Id:    aws.String(fmt.Sprintf("m%d", i+1)),
			Label: aws.String(label.String()),
			MetricStat: &cloudwatch.MetricStat{
				Metric: metric,
				Period: aws.Int64(60),
				Stat:   aws.String(stat),
			},
		})
		logrus.WithFields(logrus.Fields{
			"id":     fmt.Sprintf("m%d", i+1),
			"label":  label.String(),
			"metric": metric.GoString(),
			"stat":   stat,
		}).Debug("new metric data query")
	}
	return ret, nil
}

func interfaceToString(in interface{}) string {
	if s, ok := in.(string); ok {
		return s
	}
	return fmt.Sprintf("%s", in)
}

func setDefault(ret, last *string) {
	if *ret == "." {
		*ret = *last
	} else {
		*last = *ret
	}
}
