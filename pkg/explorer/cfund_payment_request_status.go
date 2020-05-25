package explorer

import "log"

type PaymentRequestStatus struct {
	State  uint
	Status string
}

var (
	PaymentRequestPending  = PaymentRequestStatus{0, "pending"}
	PaymentRequestAccepted = PaymentRequestStatus{1, "accepted"}
	PaymentRequestRejected = PaymentRequestStatus{2, "rejected"}
	PaymentRequestExpired  = PaymentRequestStatus{4, "expired"}
	PaymentRequestPaid     = PaymentRequestStatus{6, "paid"}
)

var paymentRequestStatus = [5]PaymentRequestStatus{
	PaymentRequestPending,
	PaymentRequestAccepted,
	PaymentRequestRejected,
	PaymentRequestExpired,
	PaymentRequestPaid,
}

//noinspection GoUnreachableCode
func GetPaymentRequestStatusByState(state uint) PaymentRequestStatus {
	for idx := range paymentRequestStatus {
		if paymentRequestStatus[idx].State == state {
			return paymentRequestStatus[idx]
		}
	}

	log.Fatal("PaymentRequestStatus state does not exist", state)
	panic(0)
}

//noinspection GoUnreachableCode
func GetPaymentRequestStatusByStatus(status string) PaymentRequestStatus {
	for idx := range paymentRequestStatus {
		if paymentRequestStatus[idx].Status == status {
			return paymentRequestStatus[idx]
		}
	}

	log.Fatal("PaymentRequestStatus status does not exist", status)
	panic(0)
}

func IsPaymentRequestStatusValid(status string) bool {
	for idx := range paymentRequestStatus {
		if paymentRequestStatus[idx].Status == status {
			return true
		}
	}
	return false
}

func IsPaymentRequestStateValid(state uint) bool {
	for idx := range paymentRequestStatus {
		if paymentRequestStatus[idx].State == state {
			return true
		}
	}
	return false
}
