package api

import (
	"github.com/clouway/go-epay/pkg/epay"
)

type checkType string

const (
	check   checkType = "CHECK"
	billing checkType = "BILLING"

	// StatusSuccess indicates the success of payment operation
	StatusSuccess string = "00"
	// StatusSubscriberNotFound indicates that unknown subscriber was requested
	StatusSubscriberNotFound string = "14"
	// StatusNoDuties indicates that subscriber has no dutiies
	StatusNoDuties string = "62"
	// StatusTemporaryNotAvailable is indicating that service is temporary not available
	StatusTemporaryNotAvailable string = "80"
	// StatusBadChecksum is indicating that received request is with bad checksum
	StatusBadChecksum string = "93"
	// StatusAlreadyPaid is indicating the duty is alredy paid
	StatusAlreadyPaid string = "94"
	// StatusCommonError indicates a common error
	StatusCommonError string = "96"

	// EPAY payment source
	EPAY epay.PaymentSource = "EPAY"
)
