package protocol

import "encoding/json"

// InquireMessage is the per-frame state sent by the server.
type InquireMessage struct {
	MatchID       string          `json:"matchId"`
	RulesVersion  string          `json:"rulesVersion"`
	Round         int             `json:"round"`
	Tick          int             `json:"tick"`
	Phase         string          `json:"phase"`
	Players       []PlayerState   `json:"players"`
	Nodes         []NodeState     `json:"nodes"`
	Edges         []EdgeDef       `json:"edges"`
	Weather       Weather         `json:"weather"`
	Tasks         []Task          `json:"tasks"`
	Bounties      []Bounty        `json:"bounties"`
	Contests      []Contest       `json:"contests"`
	Events        []Event         `json:"events"`
	ActionResults []ActionResult  `json:"actionResults"`
	ScorePreview  map[string]int  `json:"scorePreview"`
	Debug         json.RawMessage `json:"debug,omitempty"`
}

// Phase constants.
const (
	PhaseNormal = "NORMAL"
	PhaseRush   = "RUSH"
	PhaseEnded  = "ENDED"
)

// PlayerState is inquire.players[] (Appendix B).
type PlayerState struct {
	PlayerID             int             `json:"playerId"`
	Camp                 int             `json:"camp"`
	TeamID               string          `json:"teamId"`
	PlayerName           string          `json:"playerName"`
	Online               bool            `json:"online"`
	State                string          `json:"state"`
	CurrentNodeID        string          `json:"currentNodeId"`
	NextNodeID           string          `json:"nextNodeId,omitempty"`
	RouteEdgeID          string          `json:"routeEdgeId,omitempty"`
	RouteType            string          `json:"routeType,omitempty"`
	MoveDirection        string          `json:"moveDirection"`
	MoveProgress         float64         `json:"moveProgress"`
	MoveProgressRound    int             `json:"moveProgressRound"`
	CurrentEdgeCost      int             `json:"currentEdgeCost"`
	EdgeProgressPermille int             `json:"edgeProgressPermille"`
	EdgeProgressMs       int             `json:"edgeProgressMs"`
	EdgeTotalMs          int             `json:"edgeTotalMs"`
	Freshness            float64         `json:"freshness"`
	GoodFruit            int             `json:"goodFruit"`
	FrozenGoodFruit      int             `json:"frozenGoodFruit"`
	BadFruit             int             `json:"badFruit"`
	SquadAvailable       int             `json:"squadAvailable"`
	SquadInFlight        int             `json:"squadInFlight"`
	GuardActionPoint     int             `json:"guardActionPoint"`
	Verified             bool            `json:"verified"`
	Delivered            bool            `json:"delivered"`
	Retired              bool            `json:"retired"`
	RetiredRound         int             `json:"retiredRound"`
	MissingActionRounds  int             `json:"missingActionRounds"`
	IllegalActionCount   int             `json:"illegalActionCount"`
	PenaltyScore         int             `json:"penaltyScore"`
	BreakOrderReady      bool            `json:"breakOrderReady"`
	RushTacticUsedCount  int             `json:"rushTacticUsedCount"`
	Buffs                []Buff          `json:"buffs"`
	CurrentProcess       *CurrentProcess `json:"currentProcess,omitempty"`
	Resources            map[string]int  `json:"resources"`
	TotalScore           int             `json:"totalScore"`
	TaskScore            int             `json:"taskScore"`
	BountyScore          int             `json:"bountyScore"`
	ScoreDetail          ScoreDetail     `json:"scoreDetail"`
}

// Player state constants (任务书 3.1).
const (
	StateIdle          = "IDLE"
	StateMoving        = "MOVING"
	StateWaiting       = "WAITING"
	StateProcessing    = "PROCESSING"
	StateContesting    = "CONTESTING"
	StateResting       = "RESTING"
	StateForcedPassing = "FORCED_PASSING"
	StateVerifying     = "VERIFYING"
	StateCostBankrupt  = "COST_BANKRUPT"
	StateDelivered     = "DELIVERED"
	StateRetired       = "RETIRED"
)

// ScoreDetail is the score breakdown.
type ScoreDetail struct {
	Delivery   int     `json:"delivery"`
	GoodFruit  int     `json:"goodFruit"`
	Freshness  float64 `json:"freshness"`
	Time       int     `json:"time"`
	Tasks      int     `json:"tasks"`
	Bounty     int     `json:"bounty"`
	Penalty    int     `json:"penalty"`
	Total      int     `json:"total"`
}

// Buff is inquire.players[].buffs[].
type Buff struct {
	Type                string  `json:"type"`
	RemainingRound      int     `json:"remainingRound"`
	MoveMultiplier      float64 `json:"moveMultiplier,omitempty"`
	FreshnessMultiplier float64 `json:"freshnessMultiplier,omitempty"`
}

// CurrentProcess is inquire.players[].currentProcess.
type CurrentProcess struct {
	Action         string  `json:"action"`
	ObjectKey      string  `json:"objectKey,omitempty"`
	TargetNodeID   string  `json:"targetNodeId,omitempty"`
	TaskID         string  `json:"taskId,omitempty"`
	ResourceType   string  `json:"resourceType,omitempty"`
	Type           string  `json:"type,omitempty"`
	StartedRound   int     `json:"startedRound,omitempty"`
	TotalRound     int     `json:"totalRound,omitempty"`
	RemainRound    int     `json:"remainRound,omitempty"`
	RemainingRound int     `json:"remainingRound,omitempty"`
	Progress       float64 `json:"progress,omitempty"`
}

// NodeState is inquire.nodes[] (Appendix C).
type NodeState struct {
	NodeID               string           `json:"nodeId"`
	Name                 string           `json:"name"`
	X                    int              `json:"x"`
	Y                    int              `json:"y"`
	NodeType             string           `json:"nodeType"`
	ProcessType          string           `json:"processType,omitempty"`
	ProcessRound         int              `json:"processRound"`
	Start                bool             `json:"start"`
	Terminal             bool             `json:"terminal"`
	Visible              bool             `json:"visible"`
	Guard                *GuardState      `json:"guard,omitempty"`
	ResourceVisible      bool             `json:"resourceVisible"`
	ResourceStock        map[string]int   `json:"resourceStock"`
	Scouted              []ScoutMarker    `json:"scouted"`
	EffectiveCombatCount int              `json:"effectiveCombatCount"`
	GuardBlockCount      int              `json:"guardBlockCount"`
	KeyPassCombatCount   int              `json:"keyPassCombatCount"`
	HasObstacle          bool             `json:"hasObstacle"`
	ObstacleType         string           `json:"obstacleType,omitempty"`
	ObstacleResidue      *ObstacleResidue `json:"obstacleResidue,omitempty"`
	CanWindow            bool             `json:"canWindow"`
}

// GuardState is inquire.nodes[].guard.
type GuardState struct {
	OwnerTeamID    string `json:"ownerTeamId,omitempty"`
	Defense        int    `json:"defense"`
	InitialDefense int    `json:"initialDefense,omitempty"`
	MaxDefense     int    `json:"maxDefense,omitempty"`
	CompleteRound  int    `json:"completeRound,omitempty"`
	AgeRound       int    `json:"ageRound,omitempty"`
	Active         bool   `json:"active"`
}

// ScoutMarker is inquire.nodes[].scouted[].
type ScoutMarker struct {
	TeamID             string `json:"teamId"`
	RemainRound        int    `json:"remainRound"`
	ProcessReduceRound int    `json:"processReduceRound"`
	RemainingTriggers  int    `json:"remainingTriggers"`
}

// ObstacleResidue is inquire.nodes[].obstacleResidue.
type ObstacleResidue struct {
	ClearedByPlayerID int    `json:"clearedByPlayerId,omitempty"`
	ClearedByTeamID   string `json:"clearedByTeamId,omitempty"`
	ClearRound        int    `json:"clearRound,omitempty"`
	UntilRound        int    `json:"untilRound,omitempty"`
	RemainRound       int    `json:"remainRound,omitempty"`
	TaxRound          int    `json:"taxRound,omitempty"`
}

// Weather is inquire.weather (Appendix D).
type Weather struct {
	Active   []WeatherEvent `json:"active"`
	Forecast []WeatherEvent `json:"forecast"`
}

// WeatherEvent is a single weather event.
type WeatherEvent struct {
	WeatherID     string `json:"weatherId,omitempty"`
	Type          string `json:"type"`
	Region        string `json:"region,omitempty"`
	RemainRound   int    `json:"remainRound,omitempty"`
	StartRound    int    `json:"startRound,omitempty"`
	DurationRound int    `json:"durationRound,omitempty"`
}

// Task is inquire.tasks[].
type Task struct {
	TaskID             string `json:"taskId"`
	TaskTemplateID     string `json:"taskTemplateId"`
	Name               string `json:"name,omitempty"`
	NodeID             string `json:"nodeId"`
	RouteBucket        string `json:"routeBucket,omitempty"`
	ProcessType        string `json:"processType,omitempty"`
	ProcessRound       int    `json:"processRound,omitempty"`
	Score              int    `json:"score"`
	RefreshRound       int    `json:"refreshRound,omitempty"`
	ExpireRound        int    `json:"expireRound,omitempty"`
	Active             bool   `json:"active"`
	Completed          bool   `json:"completed,omitempty"`
	Failed             bool   `json:"failed,omitempty"`
	FailureReason      string `json:"failureReason,omitempty"`
	OwnerPlayerID      int    `json:"ownerPlayerId,omitempty"`
	ProtectionPlayerID int    `json:"protectionPlayerId,omitempty"`
}

// Bounty is inquire.bounties[].
type Bounty struct {
	BountyID           string `json:"bountyId"`
	BountyType         string `json:"bountyType,omitempty"`
	NodeID             string `json:"nodeId"`
	OwnerTeamID        string `json:"ownerTeamId,omitempty"`
	TriggerReason      string `json:"triggerReason,omitempty"`
	TriggerRound       int    `json:"triggerRound,omitempty"`
	CooldownUntilRound int    `json:"cooldownUntilRound,omitempty"`
	RewardScore        int    `json:"rewardScore,omitempty"`
	RewardResourceType string `json:"rewardResourceType,omitempty"`
	Active             bool   `json:"active"`
	Completed          bool   `json:"completed,omitempty"`
	WinnerPlayerID     int    `json:"winnerPlayerId,omitempty"`
}

// Contest is inquire.contests[].
type Contest struct {
	ContestID               string            `json:"contestId"`
	ContestType             string            `json:"contestType"`
	TargetNodeID            string            `json:"targetNodeId,omitempty"`
	ResourceType            string            `json:"resourceType,omitempty"`
	TaskID                  string            `json:"taskId,omitempty"`
	RedPlayerID             int               `json:"redPlayerId,omitempty"`
	BluePlayerID            int               `json:"bluePlayerId,omitempty"`
	InitiatorPlayerID       int               `json:"initiatorPlayerId,omitempty"`
	InitialTimeTaxRound     int               `json:"initialTimeTaxRound,omitempty"`
	InitialBlockType        string            `json:"initialBlockType,omitempty"`
	InitialGuardOwnerTeamID string            `json:"initialGuardOwnerTeamId,omitempty"`
	InitialGuardCompleteRound int             `json:"initialGuardCompleteRound,omitempty"`
	InitialGuardTaxRound    int               `json:"initialGuardTaxRound,omitempty"`
	InitialObstacle         bool              `json:"initialObstacle,omitempty"`
	InitialObstacleType     string            `json:"initialObstacleType,omitempty"`
	InitialObstacleTaxRound int               `json:"initialObstacleTaxRound,omitempty"`
	BreakOrderCostTypes     map[string]string `json:"breakOrderCostTypes,omitempty"`
	SourceActionTypes       map[string]string `json:"sourceActionTypes,omitempty"`
	SourceTaskIDs           map[string]string `json:"sourceTaskIds,omitempty"`
	RoundIndex              int               `json:"roundIndex,omitempty"`
	TotalRounds             int               `json:"totalRounds,omitempty"`
	RedPoint                int               `json:"redPoint,omitempty"`
	BluePoint               int               `json:"bluePoint,omitempty"`
	RedCostCount            int               `json:"redCostCount,omitempty"`
	BlueCostCount           int               `json:"blueCostCount,omitempty"`
	DeadlineRound           int               `json:"deadlineRound,omitempty"`
	Resolved                bool              `json:"resolved,omitempty"`
	WinnerTeamID            string            `json:"winnerTeamId,omitempty"`
	Cards                   map[string]string `json:"cards,omitempty"`
	Status                  string            `json:"status,omitempty"`
	ObjectKey               string            `json:"objectKey,omitempty"`
	SuppressUntilRound      int               `json:"suppressUntilRound,omitempty"`
	RemainRound             int               `json:"remainRound,omitempty"`
}

// Event is inquire.events[].
type Event struct {
	EventID string          `json:"eventId,omitempty"`
	Type    string          `json:"type"`
	Round   int             `json:"round"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ActionResult is inquire.actionResults[].
type ActionResult struct {
	Round     int    `json:"round"`
	PlayerID  int    `json:"playerId"`
	Action    string `json:"action"`
	Accepted  bool   `json:"accepted"`
	Result    string `json:"result,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`
	Message   string `json:"message,omitempty"`
}
