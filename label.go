package forwarder

import (
	"fmt"
	"strings"
)

// Label is a label for metrics.
type Label struct {
	Service    string
	HostID     string
	MetricName string
}

// ParseLabel parses a lebel.
func ParseLabel(s string) (Label, error) {
	idx := strings.IndexByte(s, ':')
	switch {
	case idx <= 0:
		return Label{}, fmt.Errorf("invalid label format, service name of host id is required: %s", s)
	case idx == len(s):
		return Label{}, fmt.Errorf("invalid label format, metric name is required: %s", s)
	}

	l, name := s[:idx], s[idx+1:]

	idx = strings.IndexByte(l, '=')
	switch {
	case idx <= 0:
		return Label{}, fmt.Errorf("invalid label format, `service' or `host' is required: %s", s)
	case idx == len(s):
		return Label{}, fmt.Errorf("invalid label format, service name of host id is required: %s", s)
	}
	t, id := l[:idx], l[idx+1:]

	switch t {
	case "service":
		return Label{
			Service:    id,
			MetricName: name,
		}, nil
	case "host":
		return Label{
			HostID:     id,
			MetricName: name,
		}, nil
	}
	return Label{}, fmt.Errorf("invalid label format, unknown id name: %s", t)
}
