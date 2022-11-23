package pipelines

const (
	QUEUED     = "queued"
	PROCESSING = "processing"
	SUCCESS    = "success"
	FAILURE    = "failure"
	ERROR      = "error"
	CANCELLED  = "cancelled"
	TIMEOUT    = "timeout"
	WAITING    = "waiting"
	SKIPPED    = "skipped"
)

/*
getPipelineStatus based on pipelines reStatus code
returns respective reStatus in string format
for eq:- 4002 return success
*/
func getPipelineStatus(statusCode int) string {
	status := "NOT DEFINED"
	switch statusCode {
	case 4000:
		return QUEUED
	case 4001:
		return PROCESSING
	case 4002:
		return SUCCESS
	case 4003:
		return FAILURE
	case 4004:
		return ERROR
	case 4005:
		return WAITING
	case 4006:
		return CANCELLED
	case 4007:
		return "unstable"
	case 4008:
		return SKIPPED
	case 4009:
		return TIMEOUT
	case 4010:
		return "stopped"
	case 4011:
		return "deleted"
	case 4012:
		return "cached"
	case 4013:
		return "cancelling"
	case 4014:
		return "timingOut"
	case 4015:
		return "creating"
	case 4016:
		return "ready"
	case 4017:
		return "online"
	case 4018:
		return "offline"
	case 4019:
		return "unhealthy"
	case 4020:
		return "onlineRequested"
	case 4021:
		return "offlineRequested"
	case 4022:
		return "pendingApproval"

	}
	return status
}
