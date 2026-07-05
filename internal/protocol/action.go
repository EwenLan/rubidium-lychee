package protocol

// Action is a single entry in action.actions[].
type Action struct {
	Action         string `json:"action"`
	TargetNodeID   string `json:"targetNodeId,omitempty"`
	ResourceType   string `json:"resourceType,omitempty"`
	TaskID         string `json:"taskId,omitempty"`
	ContestID      string `json:"contestId,omitempty"`
	Card           string `json:"card,omitempty"`
	RushTactic     string `json:"rushTactic,omitempty"`
	ExtraGoodFruit int    `json:"extraGoodFruit,omitempty"`
	GoodFruit      int    `json:"goodFruit,omitempty"`
	BadFruit       int    `json:"badFruit,omitempty"`
}

// ActionMessage is the action message sent each frame.
type ActionMessage struct {
	MatchID  string   `json:"matchId"`
	Round    int      `json:"round"`
	PlayerID int      `json:"playerId"`
	Actions  []Action `json:"actions"`
}

// Action type constants.
const (
	ActionWait           = "WAIT"
	ActionMove           = "MOVE"
	ActionDeliver        = "DELIVER"
	ActionVerifyGate     = "VERIFY_GATE"
	ActionSetGuard       = "SET_GUARD"
	ActionBreakGuard     = "BREAK_GUARD"
	ActionForcedPass     = "FORCED_PASS"
	ActionClaimResource  = "CLAIM_RESOURCE"
	ActionUseResource    = "USE_RESOURCE"
	ActionProcess        = "PROCESS"
	ActionDock           = "DOCK"
	ActionClear          = "CLEAR"
	ActionClaimTask      = "CLAIM_TASK"
	ActionSquadScout     = "SQUAD_SCOUT"
	ActionSquadClear     = "SQUAD_CLEAR"
	ActionSquadReinforce = "SQUAD_REINFORCE"
	ActionSquadWeaken    = "SQUAD_WEAKEN"
	ActionWindowCard     = "WINDOW_CARD"
	ActionRushSpeed      = "RUSH_SPEED"
	ActionRushProtect    = "RUSH_PROTECT"
)

// Window card constants.
const (
	CardYanDie    = "YAN_DIE"
	CardQiangXing = "QIANG_XING"
	CardXianGong  = "XIAN_GONG"
	CardBingZheng = "BING_ZHENG"
	CardAbstain   = "ABSTAIN"
)

// Rush tactic constants.
const (
	RushTacticBreakOrder = "BREAK_ORDER"
)
