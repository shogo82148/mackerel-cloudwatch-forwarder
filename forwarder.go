package forwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	phperjson "github.com/shogo82148/go-phper-json"
	"github.com/sirupsen/logrus"
)

// Forwarder forwards metrics of AWS CloudWatch to Mackerel
type Forwarder struct {
	Config aws.Config

	APIURL string

	// APIKey is api key for the Mackerel.
	// If it empty, the MACKEREL_APIKEY environment value is used.
	// The priority is APIKey, APIKeyParameter, MACKEREL_APIKEY, and the MACKEREL_APIKEY_PARAMETER.
	APIKey string

	// APIKeyParameter is a name of AWS Systems Manager Parameter Store for the Mackerel api key.
	// If it empty, the MACKEREL_APIKEY_PARAMETER environment value is used.
	// The priority is APIKey, APIKeyParameter, MACKEREL_APIKEY, and the MACKEREL_APIKEY_PARAMETER.
	APIKeyParameter string

	// APIKeyWithDecrypt means the Mackerel API key is encrypted.
	// If it is true, the Forwarder decrypts the API key.
	// If not, the MACKEREL_APIKEY_WITH_DECRYPT environment value is used.
	APIKeyWithDecrypt bool

	mu            sync.Mutex
	svcmackerel   *MackerelClient
	svcssm        ssmiface
	svckms        kmsiface
	svccloudwatch cloudwatchiface

	muPending             sync.Mutex
	pendingServiceMetrics serviceMetricsType
	pendingHostMetrics    hostMetricsType
}

func (f *Forwarder) mackerel(ctx context.Context) (*MackerelClient, error) {
	svcssm := f.ssm()
	svckms := f.kms()
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svcmackerel != nil {
		return f.svcmackerel, nil
	}
	key, err := f.apiKey(ctx, svcssm, svckms)
	if err != nil {
		return nil, err
	}
	f.svcmackerel = NewMackerelClient(key)
	if f.APIURL != "" {
		u, err := url.Parse(f.APIURL)
		if err != nil {
			return nil, err
		}
		f.svcmackerel.BaseURL = u
	}
	return f.svcmackerel, nil
}

func (f *Forwarder) apiKey(ctx context.Context, svcssm ssmiface, svckms kmsiface) (string, error) {
	decrypt := f.APIKeyWithDecrypt
	if os.Getenv("MACKEREL_APIKEY_WITH_DECRYPT") != "" {
		decrypt = true
	}

	if key := f.APIKey; key != "" {
		if !decrypt {
			return key, nil
		}
		b, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return "", err
		}
		resp, err := svckms.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: b,
		})
		if err != nil {
			return "", err
		}
		key = string(resp.Plaintext)
		return key, nil
	}
	if f.APIKeyParameter != "" {
		resp, err := svcssm.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(f.APIKeyParameter),
			WithDecryption: aws.Bool(decrypt),
		})
		if err != nil {
			return "", err
		}
		return aws.ToString(resp.Parameter.Value), nil
	}
	if key := os.Getenv("MACKEREL_APIKEY"); key != "" {
		if !decrypt {
			return key, nil
		}
		b, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return "", err
		}
		resp, err := svckms.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: b,
		})
		if err != nil {
			return "", err
		}
		key = string(resp.Plaintext)
		return key, nil
	}
	if name := os.Getenv("MACKEREL_APIKEY_PARAMETER"); name != "" {
		resp, err := svcssm.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(name),
			WithDecryption: aws.Bool(decrypt),
		})
		if err != nil {
			return "", err
		}
		return aws.ToString(resp.Parameter.Value), nil
	}
	return "", errors.New("forwarder: api key for the mackerel is not found")
}

func (f *Forwarder) ssm() ssmiface {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svcssm == nil {
		f.svcssm = ssm.NewFromConfig(f.Config)
	}
	return f.svcssm
}

func (f *Forwarder) kms() kmsiface {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svckms == nil {
		f.svckms = kms.NewFromConfig(f.Config)
	}
	return f.svckms
}

func (f *Forwarder) cloudwatch() cloudwatchiface {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svccloudwatch == nil {
		f.svccloudwatch = cloudwatch.NewFromConfig(f.Config)
	}
	return f.svccloudwatch
}

type forwardContext struct {
	forwarder      *Forwarder
	mackerel       *MackerelClient
	start          time.Time
	end            time.Time
	serviceMetrics serviceMetricsType
	hostMetrics    hostMetricsType

	mu                   sync.Mutex
	failedServiceMetrics serviceMetricsType
	failedHostMetrics    hostMetricsType
}

// ForwardMetrics forwards metrics of AWS CloudWatch to Mackerel
func (f *Forwarder) ForwardMetrics(ctx context.Context, data json.RawMessage) error {
	// set timeout to avoid to be killed by AWS Lambda
	timeout := 50 * time.Second
	deadline, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(deadline)
		timeout -= timeout / 10
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := f.forwardMetrics(ctx, data)
	if err != nil {
		logrus.Error(err)
	}
	return err
}

func (f *Forwarder) forwardMetrics(ctx context.Context, data json.RawMessage) error {
	var query []*Query
	if err := phperjson.Unmarshal([]byte(data), &query); err != nil {
		return fmt.Errorf("forwarder: failed to parse the input: %w", err)
	}

	now := time.Now()

	client, err := f.mackerel(ctx)
	if err != nil {
		return fmt.Errorf("forwarder: failed to configure the mackerel client: %w", err)
	}

	f.muPending.Lock()
	defer f.muPending.Unlock()

	// drop old metrics
	if cnt := f.pendingHostMetrics.Drop(now.Add(-6 * time.Hour)); cnt > 0 {
		logrus.WithFields(logrus.Fields{
			"count": cnt,
		}).Warn("drop host metrics because of timeout")
	}

	// truncate to a minute.
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html#API_GetMetricData_RequestParameters
	// > For better performance, specify StartTime and EndTime values
	// > that align with the value of the metric's Period and sync up with the beginning and end of an hour.
	start := now.Truncate(time.Minute)

	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html#publishingDataPoints
	// > When you create a metric, it can take up to 2 minutes before you can retrieve statistics
	// > for the new metric using the get-metric-statistics command.
	start = start.Add(-2 * time.Minute)
	end := start.Add(time.Minute)

	fctx := &forwardContext{
		forwarder:      f,
		mackerel:       client,
		start:          start,
		end:            end,
		serviceMetrics: f.pendingServiceMetrics,
		hostMetrics:    f.pendingHostMetrics,
	}

	err = fctx.getMetricsData(ctx, query)
	// note: do not check error here.
	// because we need to publish pending metrics.

	fctx.publishMetric(ctx)
	f.pendingServiceMetrics = fctx.failedServiceMetrics
	f.pendingHostMetrics = fctx.failedHostMetrics
	return err
}

type serviceMetricsType map[string][]ServiceMetricValue

func (m *serviceMetricsType) Append(service string, v ServiceMetricValue) {
	if *m == nil {
		*m = make(serviceMetricsType)
	}
	metrics := (*m)[service]
	for i := range metrics {
		if metrics[i].Name == v.Name && metrics[i].Time == v.Time {
			// overwrite the old value.
			metrics[i] = v
			return
		}
	}

	// append the new value.
	(*m)[service] = append(metrics, v)
}

func (m *serviceMetricsType) Drop(t time.Time) int {
	if len(*m) == 0 {
		return 0
	}
	var cnt int
	unix := t.Unix()
	for service, metrics := range *m {
		// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
		mm := metrics[:0]
		for _, v := range metrics {
			if v.Time >= unix {
				mm = append(mm, v)
			} else {
				cnt++
			}
		}
		if len(mm) > 0 {
			(*m)[service] = mm
		} else {
			delete(*m, service)
		}
	}
	return cnt
}

type hostMetricsType []HostMetricValue

func (m *hostMetricsType) Append(v HostMetricValue) {
	for i := range *m {
		if (*m)[i].HostID == v.HostID && (*m)[i].Name == v.Name && (*m)[i].Time == v.Time {
			// overwrite the old value.
			(*m)[i] = v
			return
		}
	}

	// append the new value.
	*m = append(*m, v)
}

func (m *hostMetricsType) Drop(t time.Time) int {
	if len(*m) == 0 {
		return 0
	}
	var cnt int
	unix := t.Unix()

	// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	mm := (*m)[:0]
	for _, v := range *m {
		if v.Time >= unix {
			mm = append(mm, v)
		} else {
			cnt++
		}
	}
	*m = mm
	return cnt
}

// getMetricsData gets metrics data from CloudWatch Metrics.
func (fctx *forwardContext) getMetricsData(ctx context.Context, query []*Query) error {
	svc := fctx.forwarder.cloudwatch()
	metricQuery, defaults, err := ToMetricDataQuery(query)
	if err != nil {
		return err
	}
	paginator := cloudwatch.NewGetMetricDataPaginator(svc, &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(fctx.start),
		EndTime:           aws.Time(fctx.end),
		MetricDataQueries: metricQuery,
	})
	seen := make(map[string]struct{}, len(query))
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, result := range page.MetricDataResults {
			rawLabel := aws.ToString(result.Label)
			if len(result.Values) > 0 {
				seen[rawLabel] = struct{}{}
			}
			label, err := ParseLabel(rawLabel)
			if err != nil {
				return err
			}
			for i := range result.Timestamps {
				t := result.Timestamps[i]
				v := result.Values[i]
				if label.Service != "" {
					fctx.serviceMetrics.Append(label.Service, ServiceMetricValue{
						Name:  label.MetricName,
						Time:  t.Unix(),
						Value: v,
					})
				} else if label.HostID != "" {
					fctx.hostMetrics.Append(HostMetricValue{
						HostID: label.HostID,
						Name:   label.MetricName,
						Time:   t.Unix(),
						Value:  v,
					})
				}
			}
		}
	}

	for l, v := range defaults {
		if _, ok := seen[l]; ok {
			continue
		}
		label, err := ParseLabel(l)
		if err != nil {
			return err
		}
		if label.Service != "" {
			fctx.serviceMetrics.Append(label.Service, ServiceMetricValue{
				Name:  label.MetricName,
				Time:  fctx.start.Unix(),
				Value: v,
			})
		} else if label.HostID != "" {
			fctx.hostMetrics.Append(HostMetricValue{
				HostID: label.HostID,
				Name:   label.MetricName,
				Time:   fctx.start.Unix(),
				Value:  v,
			})
		}
	}
	return nil
}

func (fctx *forwardContext) publishMetric(ctx context.Context) {
	var wg sync.WaitGroup

	// publush service metrics
	for service, metrics := range fctx.serviceMetrics {
		service, metrics := service, metrics
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fctx.mackerel.PostServiceMetricValues(ctx, service, metrics)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error":   err.Error(),
					"service": service,
				}).Warn("failed to post service metrics, will retry in next minutes")

				// save metrics to retry
				fctx.mu.Lock()
				defer fctx.mu.Unlock()
				if fctx.failedServiceMetrics == nil {
					fctx.failedServiceMetrics = make(serviceMetricsType)
				}
				fctx.failedServiceMetrics[service] = append(fctx.failedServiceMetrics[service], metrics...)
			} else {
				logrus.WithFields(logrus.Fields{
					"service": service,
					"count":   len(metrics),
				}).Info("succeed to post service metrics")
			}
		}()
	}

	// publish host metrics
	if len(fctx.hostMetrics) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fctx.mackerel.PostHostMetricValues(ctx, []HostMetricValue(fctx.hostMetrics))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Warn("failed to post host metrics, will retry in next minutes")

				// save metrics to retry
				fctx.mu.Lock()
				defer fctx.mu.Unlock()
				fctx.failedHostMetrics = fctx.hostMetrics
			} else {
				logrus.WithFields(logrus.Fields{
					"count": len(fctx.hostMetrics),
				}).Info("succeed to post host metrics")
			}
		}()
	}

	wg.Wait()
}
