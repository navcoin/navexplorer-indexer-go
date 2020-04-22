package explorer

import (
	log "github.com/sirupsen/logrus"
)

type ConsultationStatus struct {
	State  uint
	Status string
}

var (
	ConsultationPending      = ConsultationStatus{0, "waiting for support"}
	ConsultationExpired      = ConsultationStatus{3, "expired"}
	ConsultationPassed       = ConsultationStatus{7, "passed"}
	ConsultationFoundSupport = ConsultationStatus{9, "found support"}
)

var consultationStatus = [4]ConsultationStatus{
	ConsultationPending,
	ConsultationExpired,
	ConsultationPassed,
	ConsultationFoundSupport,
}

//noinspection GoUnreachableCode
func GetConsultationStatusByState(state uint) ConsultationStatus {
	for idx := range consultationStatus {
		if consultationStatus[idx].State == state {
			return consultationStatus[idx]
		}
	}

	log.Fatal("ConsultationStatus state does not exist", state)
	panic(0)
}

//noinspection GoUnreachableCode
func GetConsultationStatusByStatus(status string) ConsultationStatus {
	for idx := range consultationStatus {
		if consultationStatus[idx].Status == status {
			return consultationStatus[idx]
		}
	}

	log.Fatal("ConsultationStatus status does not exist", status)
	panic(0)
}

func IsConsultationStatusValid(status string) bool {
	for idx := range consultationStatus {
		if consultationStatus[idx].Status == status {
			return true
		}
	}
	return false
}
