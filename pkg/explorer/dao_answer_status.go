package explorer

import (
	log "github.com/sirupsen/logrus"
)

type AnswerStatus struct {
	State  uint
	Status string
}

var (
	AnswerPending   = AnswerStatus{0, "waiting for support"}
	AnswerSupported = AnswerStatus{1, "found support"}
	AnswerPassed    = AnswerStatus{7, "passed"}
)

var answerStatus = [3]AnswerStatus{
	AnswerPending,
	AnswerSupported,
	AnswerPassed,
}

//noinspection GoUnreachableCode
func GetAnswerStatusByState(state uint) AnswerStatus {
	for idx := range answerStatus {
		if answerStatus[idx].State == state {
			return answerStatus[idx]
		}
	}

	log.Fatal("AnswerStatus state does not exist: ", state)
	panic(0)
}

//noinspection GoUnreachableCode
func GetAnswerStatusByStatus(status string) AnswerStatus {
	for idx := range answerStatus {
		if answerStatus[idx].Status == status {
			return answerStatus[idx]
		}
	}

	log.Fatal("ConsultationStatus status does not exist: ", status)
	panic(0)
}

func IsAnswerStatusValid(status string) bool {
	for idx := range answerStatus {
		if answerStatus[idx].Status == status {
			return true
		}
	}
	return false
}
