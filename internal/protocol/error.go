package protocol

// ErrorMessage is sent by the server when a client packet fails to enter
// the rule settlement (bad framing, unknown msg_name, mismatched matchId, etc.).
// Rule-level rejections are NOT sent here; they appear in the next frame's
// events[] and actionResults[].
type ErrorMessage struct {
	Round     int    `json:"round,omitempty"`
	PlayerID  int    `json:"playerId,omitempty"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message,omitempty"`
}

// Immediate error codes (returned as msg_name:error).
const (
	ErrInvalidLengthPrefix   = "INVALID_LENGTH_PREFIX"
	ErrInvalidJSON           = "INVALID_JSON"
	ErrInvalidActionType     = "INVALID_ACTION_TYPE"
	ErrActionRejected        = "ACTION_REJECTED"
	ErrMatchIDMismatch       = "MATCH_ID_MISMATCH"
	ErrActionTooLate         = "ACTION_TOO_LATE"
	ErrDuplicateAction       = "DUPLICATE_ACTION"
	ErrPlayerAddressMismatch = "PLAYER_ADDRESS_MISMATCH"
	ErrPlayerNotAllowed      = "PLAYER_NOT_ALLOWED"
	ErrMatchAlreadyStarted   = "MATCH_ALREADY_STARTED"
	ErrPlayerLimitExceeded   = "PLAYER_LIMIT_EXCEEDED"
)

// Rule rejection error codes (returned in next-frame events[]/actionResults[]).
const (
	ErrInvalidActionConflict    = "INVALID_ACTION_CONFLICT"
	ErrParamOutOfRange          = "PARAM_OUT_OF_RANGE"
	ErrMovingActionForbidden    = "MOVING_ACTION_FORBIDDEN"
	ErrRestingActionForbidden   = "RESTING_ACTION_FORBIDDEN"
	ErrSafeZoneForbidden        = "SAFE_ZONE_FORBIDDEN"
	ErrProcessRequired          = "PROCESS_REQUIRED"
	ErrProcessNotAvailable      = "PROCESS_NOT_AVAILABLE"
	ErrNotAtTargetNode          = "NOT_AT_TARGET_NODE"
	ErrMoveMissingTarget        = "MOVE_MISSING_TARGET"
	ErrMoveEdgeNotFound         = "MOVE_EDGE_NOT_FOUND"
	ErrMoveBlockedByGuard       = "MOVE_BLOCKED_BY_GUARD"
	ErrTargetNotFound           = "TARGET_NOT_FOUND"
	ErrTargetNotReachable       = "TARGET_NOT_REACHABLE"
	ErrResourceNotEnough        = "RESOURCE_NOT_ENOUGH"
	ErrResourceNotUsable        = "RESOURCE_NOT_USABLE"
	ErrTaskNotFound             = "TASK_NOT_FOUND"
	ErrTaskProtected            = "TASK_PROTECTED"
	ErrTaskRequirementNotMet    = "TASK_REQUIREMENT_NOT_MET"
	ErrTaskExpired              = "TASK_EXPIRED"
	ErrObjectBusy               = "OBJECT_BUSY"
	ErrWindowDrawRetryLimit     = "WINDOW_DRAW_RETRY_LIMIT"
	ErrVerifyRequired           = "VERIFY_REQUIRED"
	ErrAlreadyVerified          = "ALREADY_VERIFIED"
	ErrDeliverNotAtTerminal     = "DELIVER_NOT_AT_TERMINAL"
	ErrDeliverNotVerified       = "DELIVER_NOT_VERIFIED"
	ErrDeliverRequirementNotMet = "DELIVER_REQUIREMENT_NOT_MET"
	ErrAlreadyDelivered         = "ALREADY_DELIVERED"
	ErrRushTacticInvalidBinding = "RUSH_TACTIC_INVALID_BINDING"
	ErrHorseBuffConflict        = "HORSE_BUFF_CONFLICT"
	ErrForcedPassRepeat         = "FORCED_PASS_REPEAT"
	ErrObstacleNotFound         = "OBSTACLE_NOT_FOUND"
)
