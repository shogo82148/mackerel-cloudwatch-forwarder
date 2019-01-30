package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var defaultBaseURL *url.URL

func init() {
	var err error
	defaultBaseURL, err = url.Parse("https://api.mackerelio.com/")
	if err != nil {
		panic(err)
	}
}

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
	BaseURL    *url.URL
	APIKey     string
	UserAgent  string
	HTTPClient *http.Client
}

func (c *MackerelClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *MackerelClient) urlfor(path string) string {
	base := c.BaseURL
	if base == nil {
		base = defaultBaseURL
	}

	// shallow copy
	u := new(url.URL)
	*u = *base

	u.Path = path
	return u.String()
}

func (c *MackerelClient) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	u := c.urlfor(path)
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("X-Api-Key", c.APIKey)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	} else {
		agent := fmt.Sprintf("mackerel-cloudwatch-forwarder/%s", Version)
		req.Header.Set("User-Agent", agent)
	}

	return req, nil
}

func (c *MackerelClient) postJSON(ctx context.Context, path string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return handleError(resp)
	}

	io.Copy(ioutil.Discard, resp.Body)

	return nil
}

// Error is an error from the Mackerel.
type Error struct {
	StatusCode int
	Message    string
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d, %s", e.StatusCode, e.Message)
}

func handleError(resp *http.Response) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}
}

// PostServiceMetricValues posts service metrics.
func (c *MackerelClient) PostServiceMetricValues(ctx context.Context, serviceName string, values []*ServiceMetricValue) error {
	return c.postJSON(ctx, fmt.Sprintf("api/v0/services/%s/tsdb", serviceName), values)
}

// PostHostMetricValues posts host metrics.
func (c *MackerelClient) PostHostMetricValues(ctx context.Context, values []*HostMetricValue) error {
	return c.postJSON(ctx, "api/v0/tsdb", values)
}
