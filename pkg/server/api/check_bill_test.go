package api

import (
	"testing"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/google/go-cmp/cmp"
)

func TestSuccessResponse(t *testing.T) {
	cases := []struct {
		name         string
		subscriberID string
		customerName string
		items        []epay.Item
		cents        int
		want         *DutyResponse
	}{
		{
			name:         "short desc is limited",
			subscriberID: "1234567",
			customerName: "ЕРДОАН ЕФРАИМОВ ЕФРАИМОВ",
			cents:        100,
			want: &DutyResponse{
				Status:    "00",
				IDN:       "1234567",
				ShortDesc: "ЕРДОАН ЕФРАИМОВ ЕФРАИ",
				LongDesc:  "Клиент: ЕРДОАН ЕФРАИМОВ ЕФРАИМОВ, Абонатен Номер: 1234567, Детайли: ",
				Amount:    100,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := successResponse(c.subscriberID, c.customerName, c.items, c.cents)

			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal("unexpected response (-want +got): ", diff)
			}
		})
	}
}
