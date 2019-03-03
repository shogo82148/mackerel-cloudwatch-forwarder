package forwarder

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
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
	svcssm        ssmiface.SSMAPI
	svckms        kmsiface.KMSAPI
	svccloudwatch cloudwatchiface.CloudWatchAPI

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
	f.svcmackerel = &MackerelClient{
		APIKey: key,
	}
	if f.APIURL != "" {
		u, err := url.Parse(f.APIURL)
		if err != nil {
			return nil, err
		}
		f.svcmackerel.BaseURL = u
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

type forwardContext struct {
	context.Context
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
func (f *Forwarder) ForwardMetrics(ctx context.Context, query []cloudwatch.MetricDataQuery) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	client, err := f.mackerel(ctx)
	if err != nil {
		return err
	}

	f.muPending.Lock()
	defer f.muPending.Unlock()

	fctx := &forwardContext{
		Context:        ctx,
		forwarder:      f,
		mackerel:       client,
		start:          now.Add(-3 * time.Minute),
		end:            now,
		serviceMetrics: f.pendingServiceMetrics,
		hostMetrics:    f.pendingHostMetrics,
	}

	err = fctx.getMetricsData(query)
	// note: do not check error here.
	// because we need to publish pending metrics.

	fctx.publishMetric()
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

// getMetricsData gets metrics data from CloudWatch Metrics.
func (fctx *forwardContext) getMetricsData(query []cloudwatch.MetricDataQuery) error {
	svc := fctx.forwarder.cloudwatch()
	in := &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(fctx.start),
		EndTime:           aws.Time(fctx.end),
		MetricDataQueries: query,
	}
	for {
		req := svc.GetMetricDataRequest(in)
		req.SetContext(fctx)
		page, err := req.Send()
		if err != nil {
			return err
		}
		for _, result := range page.MetricDataResults {
			label, err := ParseLabel(aws.StringValue(result.Label))
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
				} else {
					fctx.hostMetrics.Append(HostMetricValue{
						HostID: label.HostID,
						Name:   label.MetricName,
						Time:   t.Unix(),
						Value:  v,
					})
				}
			}
		}
		if page.NextToken == nil {
			break
		}
		in.NextToken = page.NextToken
	}
	return nil
}

func (fctx *forwardContext) publishMetric() {
	var wg sync.WaitGroup

	// publush service metrics
	for service, metrics := range fctx.serviceMetrics {
		service, metrics := service, metrics
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fctx.mackerel.PostServiceMetricValues(fctx, service, metrics)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := fctx.mackerel.PostHostMetricValues(fctx, []HostMetricValue(fctx.hostMetrics))
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

	wg.Wait()
}
