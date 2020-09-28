package explorer

type SoftForkState string

var (
	SoftForkDefined  SoftForkState = "defined"
	SoftForkStarted  SoftForkState = "started"
	SoftForkLockedIn SoftForkState = "locked_in"
	SoftForkActive   SoftForkState = "active"
	SoftForkFailed   SoftForkState = "failed"
)
