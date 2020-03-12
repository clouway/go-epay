package epay

import (
	"net/url"
	"testing"
)

func TestSum(t *testing.T) {
	ecs := "c981d4c17f7e01a71d021590e97d57c0a3da21b9"

	values := url.Values{
		"IDN":        []string{"::idn::"},
		"CHECKSUM":   []string{ecs},
		"MERCHANTID": []string{"MARCHANTID"},
	}

	cs := Checksum(values, "mysecret")

	if ecs != cs {
		t.Errorf("expected Checksum(v, secret) to be: %s", ecs)
		t.Errorf("                           but was: %s", cs)
	}
}
