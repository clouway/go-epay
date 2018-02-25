package epay

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		message string
		exp     request
	}{
		{"XTYPE=QBN\nIDN=123\nTID=TID123\n", request{Type: "QBN", CustomerID: "123", TransactionID: "TID123"}},
		{"XTYPE=QBN\nIDN=321\nTID=TID321\n", request{Type: "QBN", CustomerID: "321", TransactionID: "TID321"}},
		{"XTYPE=QBC\nIDN=321\nTID=TID321\nAMOUNT=120\n", request{Type: "QBC", CustomerID: "321", TransactionID: "TID321", Amount: 120}},
		{"", request{}},
	}

	for _, c := range cases {
		buf := bytes.NewBufferString(c.message)
		cmd, _ := parseRequest(buf)

		if !reflect.DeepEqual(*cmd, c.exp) {
			t.Errorf("expected: %v", c.exp)
			t.Errorf("     got: %v", cmd)
		}
	}
}
