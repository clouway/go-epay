package epay

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
)

// Checksum calculates the Values checksum using the epay specific format.
func Checksum(q url.Values, secret string) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		if k == "CHECKSUM" {
			continue
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)
	message := ""
	for _, k := range keys {
		message += fmt.Sprintf("%s%s\n", k, q[k][0])
	}

	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
