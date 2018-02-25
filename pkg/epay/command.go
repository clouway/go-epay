package epay

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseRequest(r io.Reader) (*request, error) {
	buf := make([]byte, 512)
	pairs := make(map[string]string)
	for {
		n, err := r.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		content := string(buf)
		if !strings.Contains(content, "=") {
			return nil, fmt.Errorf("could not find command separator")
		}

		lines := strings.Split(content, "\n")
		for _, line := range lines {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				pairs[parts[0]] = parts[1]
			}
		}
	}

	c := &request{Type: pairs["XTYPE"], CustomerID: pairs["IDN"], TransactionID: pairs["TID"]}

	if amount, err := strconv.Atoi(pairs["AMOUNT"]); err == nil {
		c.Amount = amount
	}

	return c, nil

}

// IsForBillCheck determines whether it's a bill check request. Returns true if it's
// a bill check request and false in other case
func (c *request) IsForBillCheck() bool {
	return strings.EqualFold(c.Type, "QBN")
}

// IsForPayment checks whether command is for payment processing
func (c *request) IsForPayment() bool {
	return strings.EqualFold(c.Type, "QBC")
}
