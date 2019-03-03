package forwarder

import (
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	testcases := []struct {
		in    string
		out   Label
		valid bool
	}{
		{
			in: "service=prod:foo.bar.baz",
			out: Label{
				Service:    "prod",
				MetricName: "foo.bar.baz",
			},
			valid: true,
		},
		{
			in: "host=abcdefg:boo.foo.uoo",
			out: Label{
				HostID:     "abcdefg",
				MetricName: "boo.foo.uoo",
			},
			valid: true,
		},
		{
			in: "",
		},
		{
			in: "zzz:foo.bar.baz",
		},
		{
			in: "zzz=goo:foo.bar.baz",
		},
		{
			in: "=goo:foo.bar.baz",
		},
		{
			in: "goo=:foo.bar.baz",
		},
		{
			in: "foo.bar.baz",
		},
		{
			in: ":foo.bar.baz",
		},
		{
			in: "foo.bar.baz:",
		},
	}

	for i, s := range testcases {
		out, err := ParseLabel(s.in)
		if s.valid {
			if err != nil {
				t.Errorf("no.%d: error: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(out, s.out) {
				t.Errorf("no.%d: want %s, got %s", i, s.out, out)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("no.%d: want error, got nil", i)
			}
		}
	}
}
