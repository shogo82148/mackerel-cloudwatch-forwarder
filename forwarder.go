package forwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
)

// Forwarder forwards metrics of AWS CloudWatch to Mackerel
type Forwarder struct {
	Config aws.Config

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

	// number of concurrent events, access atomically
	events int64

	mu            sync.Mutex
	ch            chan struct{}
	svcmackerel   *MackerelClient
	svcssm        ssmiface.SSMAPI
	svckms        kmsiface.KMSAPI
	svccloudwatch cloudwatchiface.CloudWatchAPI
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
	f.svcmackerel = &MackerelClient{
		APIKey: key,
	}
	return f.svcmackerel, nil
}

func (f *Forwarder) apiKey(ctx context.Context, svcssm ssmiface.SSMAPI, svckms kmsiface.KMSAPI) (string, error) {
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
		req := svckms.DecryptRequest(&kms.DecryptInput{
			CiphertextBlob: b,
		})
		req.SetContext(ctx)
		resp, err := req.Send()
		if err != nil {
			return "", err
		}
		key = string(resp.Plaintext)
		return key, nil
	}
	if f.APIKeyParameter != "" {
		req := svcssm.GetParameterRequest(&ssm.GetParameterInput{
			Name:           aws.String(f.APIKeyParameter),
			WithDecryption: aws.Bool(decrypt),
		})
		req.SetContext(ctx)
		resp, err := req.Send()
		if err != nil {
			return "", err
		}
		return aws.StringValue(resp.Parameter.Value), nil
	}
	if key := os.Getenv("MACKEREL_APIKEY"); key != "" {
		if !decrypt {
			return key, nil
		}
		b, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return "", err
		}
		req := svckms.DecryptRequest(&kms.DecryptInput{
			CiphertextBlob: b,
		})
		req.SetContext(ctx)
		resp, err := req.Send()
		if err != nil {
			return "", err
		}
		key = string(resp.Plaintext)
		return key, nil
	}
	if name := os.Getenv("MACKEREL_APIKEY_PARAMETER"); name != "" {
		req := svcssm.GetParameterRequest(&ssm.GetParameterInput{
			Name:           aws.String(name),
			WithDecryption: aws.Bool(decrypt),
		})
		req.SetContext(ctx)
		resp, err := req.Send()
		if err != nil {
			return "", err
		}
		return aws.StringValue(resp.Parameter.Value), nil
	}
	return "", errors.New("forwarder: api key for the mackerel is not found")
}

func (f *Forwarder) ssm() ssmiface.SSMAPI {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svcssm == nil {
		f.svcssm = ssm.New(f.Config)
	}
	return f.svcssm
}

func (f *Forwarder) kms() kmsiface.KMSAPI {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svckms == nil {
		f.svckms = kms.New(f.Config)
	}
	return f.svckms
}

func (f *Forwarder) cloudwatch() cloudwatchiface.CloudWatchAPI {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.svccloudwatch == nil {
		f.svccloudwatch = cloudwatch.New(f.Config)
	}
	return f.svccloudwatch
}

func (f *Forwarder) chfinished() chan struct{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ch == nil {
		f.ch = make(chan struct{}, 1)
	}
	return f.ch
}

type nowKey struct{}

func withTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, nowKey{}, t)
}

func now(ctx context.Context) time.Time {
	return ctx.Value(nowKey{}).(time.Time)
}

// ForwardMetrics forwards metrics of AWS CloudWatch to Mackerel
func (f *Forwarder) ForwardMetrics(ctx context.Context, event ForwardMetricsEvent) error {
	now := time.Now()
	timestamp := now.Format(time.RFC3339Nano)
	ctx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	atomic.AddInt64(&f.events, 1)
	chfinished := f.chfinished()
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 7*24*time.Hour)
		defer cancel()
		ctx = withTime(ctx, now)
		err := f.forwardMetrics(ctx, timestamp, event)
		if err != nil {
			log.Printf("metric[%s]: finished %v", timestamp, err)
		} else {
			log.Printf("metric[%s]: finished", timestamp)
		}
		chfinished <- struct{}{}
	}()

	for {
		select {
		case <-chfinished:
			events := atomic.AddInt64(&f.events, -1)
			if events == 0 {
				return nil
			}
		case <-ctx.Done():
			log.Printf("metric[%s]: %v", timestamp, ctx.Err())
			return nil
		}
	}
}

func (f *Forwarder) forwardMetrics(ctx context.Context, timestamp string, event ForwardMetricsEvent) error {
	var errCount int64

	// forward service metrics
	var wg sync.WaitGroup
	for _, def := range event.ServiceMetrics {
		def := def
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f.forwardServiceMetric(ctx, def); err != nil {
				log.Printf("metric[%s]: %v", timestamp, err)
				atomic.AddInt64(&errCount, 1)
			}
		}()
	}

	// forward host metrics
	for _, def := range event.HostMetrics {
		def := def
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f.forwardHostMetric(ctx, def); err != nil {
				log.Printf("metric[%s]: %v", timestamp, err)
				atomic.AddInt64(&errCount, 1)
			}
		}()
	}

	wg.Wait()
	cnt := atomic.LoadInt64(&errCount)
	if cnt != 0 {
		return fmt.Errorf("%d error(s)", cnt)
	}
	return nil
}

func (f *Forwarder) forwardServiceMetric(ctx context.Context, def ServiceMetricDefinition) error {
	c, err := f.mackerel(ctx)
	if err != nil {
		return err
	}

	m, err := f.GetServiceMetric(ctx, def)
	if err != nil {
		return err
	}
	if err := c.PostServiceMetricValues(ctx, def.Service, m); err != nil {
		return err
	}
	return nil
}

func (f *Forwarder) forwardHostMetric(ctx context.Context, def HostMetricDefinition) error {
	c, err := f.mackerel(ctx)
	if err != nil {
		return err
	}

	m, err := f.GetHostMetric(ctx, def)
	if err != nil {
		return err
	}
	if err := c.PostHostMetricValues(ctx, m); err != nil {
		return err
	}
	return nil
}

// GetServiceMetric gets service metrics from CloudWatch.
func (f *Forwarder) GetServiceMetric(ctx context.Context, def ServiceMetricDefinition) ([]*ServiceMetricValue, error) {
	now := now(ctx)
	prev := now.Add(-2 * time.Minute) // 2 min (to fetch at least 1 data-point)

	template := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: []cloudwatch.Dimension{},
		StartTime:  aws.Time(prev),
		EndTime:    aws.Time(now),
		Period:     aws.Int64(60),
	}
	if err := setStatistics(template, def.Stat); err != nil {
		return nil, err
	}

	input, err := ParseMetric(template, def.Metric)
	if err != nil {
		return nil, err
	}
	req := f.cloudwatch().GetMetricStatisticsRequest(input)
	req.SetContext(ctx)
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}

	ret := make([]*ServiceMetricValue, 0, len(resp.Datapoints))
	for _, p := range resp.Datapoints {
		ret = append(ret, &ServiceMetricValue{
			Name:  def.Name,
			Time:  p.Timestamp.Unix(),
			Value: getStatistics(p, def.Stat),
		})
	}
	return ret, nil
}

// GetHostMetric gets service metrics from CloudWatch.
func (f *Forwarder) GetHostMetric(ctx context.Context, def HostMetricDefinition) ([]*HostMetricValue, error) {
	now := now(ctx)
	prev := now.Add(-2 * time.Minute) // 2 min (to fetch at least 1 data-point)

	template := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: []cloudwatch.Dimension{},
		StartTime:  aws.Time(prev),
		EndTime:    aws.Time(now),
		Period:     aws.Int64(60),
	}
	if err := setStatistics(template, def.Stat); err != nil {
		return nil, err
	}

	input, err := ParseMetric(template, def.Metric)
	if err != nil {
		return nil, err
	}
	req := f.cloudwatch().GetMetricStatisticsRequest(input)
	req.SetContext(ctx)
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}

	ret := make([]*HostMetricValue, 0, len(resp.Datapoints))
	for _, p := range resp.Datapoints {
		ret = append(ret, &HostMetricValue{
			HostID: def.HostID,
			Name:   def.Name,
			Time:   p.Timestamp.Unix(),
			Value:  getStatistics(p, def.Stat),
		})
	}
	return ret, nil
}

func setStatistics(input *cloudwatch.GetMetricStatisticsInput, stat string) error {
	if strings.HasPrefix(stat, "p") {
		// it looks like percentile statistics
		input.ExtendedStatistics = []string{stat} // XXX: need validations?
		return nil
	}
	// otherwise, maybe normal statistics.
	input.Statistics = []cloudwatch.Statistic{cloudwatch.Statistic(stat)}
	return nil
}

func getStatistics(p cloudwatch.Datapoint, stat string) float64 {
	if strings.HasPrefix(stat, "p") {
		// it looks like percentile statistics
		return p.ExtendedStatistics[stat]
	}
	// otherwise, maybe normal statistics.
	// See https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricStatistics.html
	switch stat {
	case "SampleCount":
		return aws.Float64Value(p.SampleCount)
	case "Average":
		return aws.Float64Value(p.Average)
	case "Sum":
		return aws.Float64Value(p.Sum)
	case "Minimum":
		return aws.Float64Value(p.Minimum)
	case "Maximum":
		return aws.Float64Value(p.Maximum)
	}
	return 0
}

// ForwardMetricsEvent is an event of ForwardMetrics.
type ForwardMetricsEvent struct {
	ServiceMetrics []ServiceMetricDefinition `json:"service_metrics"`
	HostMetrics    []HostMetricDefinition    `json:"host_metrics"`
}

// ServiceMetricDefinition is a definition for converting a metric of AWS CloudWatch to Mackerel's Service Metrics.
// https://mackerel.io/api-docs/entry/service-metrics
type ServiceMetricDefinition struct {
	Service string      `json:"service"`
	Name    string      `json:"name"`
	Metric  interface{} `json:"metric"`
	Stat    string      `json:"stat"`
}

// HostMetricDefinition is a definition for converting a metric of AWS CloudWatch to Mackerel's Host Metrics.
// https://mackerel.io/api-docs/entry/host-metrics
type HostMetricDefinition struct {
	HostID string      `json:"hostId"`
	Name   string      `json:"name"`
	Metric interface{} `json:"metric"`
	Stat   string      `json:"stat"`
}

// ParseMetric parses the metrics definitions.
// See https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/CloudWatch-Dashboard-Body-Structure.html#CloudWatch-Dashboard-Properties-Metrics-Array-Format
// The rendering properties object will be ignored.
func ParseMetric(template *cloudwatch.GetMetricStatisticsInput, def interface{}) (*cloudwatch.GetMetricStatisticsInput, error) {
	var ret cloudwatch.GetMetricStatisticsInput
	ret = *template

	var array []interface{}
	switch def := def.(type) {
	case []interface{}:
		array = def
	case []string:
		array = make([]interface{}, 0, len(def))
		for _, v := range def {
			array = append(array, v)
		}
	case string:
		if err := json.Unmarshal([]byte(def), &array); err != nil {
			return nil, err
		}
	case []byte:
		if err := json.Unmarshal(def, &array); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("forwarder: type of metrics definition is invalid: %T", def)
	}

	if len(array) < 2 {
		return nil, errors.New("forwarder: Namespace and MetricName are required")
	}

	namespace, ok := array[0].(string)
	if !ok {
		return nil, fmt.Errorf("forwarder: invalid type of Namespace: %T", array[0])
	}
	ret.Namespace = aws.String(namespace)

	metricName, ok := array[1].(string)
	if !ok {
		return nil, fmt.Errorf("forwarder: invalid type of MetricName: %T", array[1])
	}
	ret.MetricName = aws.String(metricName)

	dimensions := []cloudwatch.Dimension{}
	for i := 2; i+1 < len(array); i += 2 {
		name, ok := array[i].(string)
		if !ok {
			return nil, fmt.Errorf("forwarder: invalid type of DimensionName: %T", array[i])
		}
		value, ok := array[i+1].(string)
		if !ok {
			return nil, fmt.Errorf("forwarder: invalid type of DimensionValue: %T", array[i+1])
		}
		dimensions = append(dimensions, cloudwatch.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}
	ret.Dimensions = dimensions

	return &ret, nil
}
