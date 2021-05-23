package forwarder

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestPostServiceMetricValues(t *testing.T) {
	ch := make(chan interface{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: want %s, got %s", http.MethodPost, r.Method)
		}
		if want, got := "api-token", r.Header.Get("X-Api-Key"); want != got {
			t.Errorf("unexpected api token: want %q, got %q", want, got)
		}
		if want, got := "/api/v0/services/awesome-service/tsdb", r.URL.Path; want != got {
			t.Errorf("unexpected path: want %q, got %q", want, got)
		}

		var body interface{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		ch <- body
		rw.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	client := NewMackerelClient("api-token")
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	err = client.PostServiceMetricValues(context.Background(), "awesome-service", []ServiceMetricValue{
		{
			Name:  "metric.sum",
			Time:  1234567890.0,
			Value: 123.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var got interface{}
	select {
	case got = <-ch:
	default:
		t.Fatal("api is not called")
	}
	want := []interface{}{
		map[string]interface{}{
			"name":  "metric.sum",
			"time":  1234567890.0,
			"value": 123.0,
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("body mismatch: (-want/+got):\n%s", diff)
	}
}

func TestPostHostMetricValues(t *testing.T) {
	ch := make(chan interface{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: want %s, got %s", http.MethodPost, r.Method)
		}
		if want, got := "api-token", r.Header.Get("X-Api-Key"); want != got {
			t.Errorf("unexpected api token: want %q, got %q", want, got)
		}
		if want, got := "/api/v0/tsdb", r.URL.Path; want != got {
			t.Errorf("unexpected path: want %q, got %q", want, got)
		}

		var body interface{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		ch <- body
		rw.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	client := NewMackerelClient("api-token")
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	err = client.PostHostMetricValues(context.Background(), []HostMetricValue{
		{
			HostID: "host-abc",
			Name:   "metric.sum",
			Time:   1234567890.0,
			Value:  123.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var got interface{}
	select {
	case got = <-ch:
	default:
		t.Fatal("api is not called")
	}
	want := []interface{}{
		map[string]interface{}{
			"hostId": "host-abc",
			"name":   "metric.sum",
			"time":   1234567890.0,
			"value":  123.0,
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("body mismatch: (-want/+got):\n%s", diff)
	}
}

func TestPostServiceMetricValues_TemporaryError(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cnt := atomic.AddInt32(&count, 1)
		// api fails first time
		if cnt <= 1 {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	client := NewMackerelClient("api-token")
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	err = client.PostServiceMetricValues(context.Background(), "awesome-service", []ServiceMetricValue{
		{
			Name:  "metric.sum",
			Time:  1234567890.0,
			Value: 123.0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want, got := int32(2), atomic.LoadInt32(&count); want != got {
		t.Errorf("unexpected api call count: want %d, got %d", want, got)
	}
}

func TestPostServiceMetricValues_ClientError(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		rw.WriteHeader(http.StatusForbidden)
		io.WriteString(rw, "forbidden")
	}))
	defer ts.Close()
	client := NewMackerelClient("api-token")
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	err = client.PostServiceMetricValues(context.Background(), "awesome-service", []ServiceMetricValue{
		{
			Name:  "metric.sum",
			Time:  1234567890.0,
			Value: 123.0,
		},
	})
	if err == nil {
		t.Errorf("want error, got nil")
	}
	var merr Error
	if !errors.As(err, &merr) {
		t.Errorf("want forwader.Error type, got %T", err)
	}
	if merr.StatusCode != http.StatusForbidden {
		t.Errorf("unexpected status code: want %d, got %d", http.StatusForbidden, merr.StatusCode)
	}
	if merr.Message != "forbidden" {
		t.Errorf("unexpected message: want %q, got %q", "forbidden", merr.Message)
	}
	if want, got := int32(1), atomic.LoadInt32(&count); want != got {
		t.Errorf("unexpected api call count: want %d, got %d", want, got)
	}
}

func TestPostServiceMetricValues_Error(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		rw.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()
	client := NewMackerelClient("api-token")
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	// default retry policy takes a long time
	// so overwrite shorter settings
	client.RetryPolicy.MinDelay = 100 * time.Millisecond
	client.RetryPolicy.MaxDelay = 100 * time.Millisecond
	client.RetryPolicy.Jitter = 100 * time.Millisecond
	client.RetryPolicy.MaxCount = 10

	err = client.PostServiceMetricValues(context.Background(), "awesome-service", []ServiceMetricValue{
		{
			Name:  "metric.sum",
			Time:  1234567890.0,
			Value: 123.0,
		},
	})
	if err == nil {
		t.Errorf("want error, got nil")
	}
	var merr Error
	if !errors.As(err, &merr) {
		t.Errorf("want forwader.Error type, got %T", err)
	}
	if merr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("unexpected status code: want %d, got %d", http.StatusTooManyRequests, merr.StatusCode)
	}
	if want, got := int32(10), atomic.LoadInt32(&count); want != got {
		t.Errorf("unexpected api call count: want %d, got %d", want, got)
	}
}
