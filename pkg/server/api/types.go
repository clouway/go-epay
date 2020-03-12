package api

var (
	// ErrBadChecksum is indicating that received request is with bad checksum
	ErrBadChecksum = &DutyResponse{Status: "93"}
)

// DutyResponse is a common response which is returned from the server.
type DutyResponse struct {
	Status    string `json:"STATUS,omitempty"`
	IDN       string `json:"IDN,omitempty"`
	ShortDesc string `json:"SHORTDESC,omitempty"`
	LongDesc  string `json:"LONGDESC,omitempty"`
	Amount    int    `json:"AMOUNT,omitempty"`
	ValidTo   string `json:"VALIDTO,omitempty"`
}
