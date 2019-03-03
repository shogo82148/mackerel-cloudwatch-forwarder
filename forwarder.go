package forwarder

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
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
func (f *Forwarder) ForwardMetrics(ctx context.Context, query []cloudwatch.MetricDataQuery) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	client, err := f.mackerel(ctx)
	if err != nil {
		return err
	}

	svc := f.cloudwatch()
	in := &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(now.Add(-3 * time.Minute)),
		EndTime:           aws.Time(now),
		MetricDataQueries: query,
		MaxDatapoints:     aws.Int64(1),
	}
	for {
		log.Println(in)
		req := svc.GetMetricDataRequest(in)
		req.SetContext(ctx)
		page, err := req.Send()
		if err != nil {
			return err
		}
		log.Println(page)
		for _, result := range page.MetricDataResults {
			label, err := ParseLabel(aws.StringValue(result.Label))
			if err != nil {
				log.Println(err)
				continue
			}
			for i := range result.Timestamps {
				t := result.Timestamps[i]
				v := result.Values[i]
				if label.Service != "" {
					err := client.PostServiceMetricValues(ctx, label.Service, []*ServiceMetricValue{
						{
							Name:  label.MetricName,
							Time:  t.Unix(),
							Value: v,
						},
					})
					if err != nil {
						log.Println(err)
						continue
					}
				} else {
					err := client.PostHostMetricValues(ctx, []*HostMetricValue{
						{
							HostID: label.HostID,
							Name:   label.MetricName,
							Time:   t.Unix(),
							Value:  v,
						},
					})
					if err != nil {
						log.Println(err)
						continue
					}
				}
			}
		}
		log.Println(page.NextToken)
		if page.NextToken == nil {
			break
		}
		in.NextToken = page.NextToken
	}

	return nil
}
