package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL    = "https://api.mackerelio.com/"
	defaultUserAgent  = "mackerel-client-go"
	apiRequestTimeout = 30 * time.Second
)

// ServiceMetricValue is a value of Mackerel's service metric.
type ServiceMetricValue struct {
	Name  string  `json:"name"`
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

// HostMetricValue is a value of Mackerel's host metric.
type HostMetricValue struct {
	HostID string  `json:"hostId"`
	Name   string  `json:"name"`
	Time   int64   `json:"time"`
	Value  float64 `json:"value"`
}

// MackerelClient is a tiny client for Mackerel.
type MackerelClient struct {
	// BaseURL    *url.URL // TODO
	APIKey string
	// UserAgent  string // TODO
	// HTTPClient *http.Client // TODO
}

// PostServiceMetricValues posts service metrics.
func (c *MackerelClient) PostServiceMetricValues(ctx context.Context, serviceName string, values []*ServiceMetricValue) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%sapi/v0/services/%s/tsdb", defaultBaseURL, serviceName), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO check the status code.
	// TODO retry

	io.Copy(os.Stderr, resp.Body)

	return nil
}

// PostHostMetricValues posts host metrics.
func (c *MackerelClient) PostHostMetricValues(ctx context.Context, values []*HostMetricValue) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%sapi/v0/tsdb", defaultBaseURL), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO check the status code.
	// TODO retry

	io.Copy(os.Stderr, resp.Body)

	return nil
}
