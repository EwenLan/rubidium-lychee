package protocol

// OverMessage is the final message sent when the match ends.
type OverMessage struct {
	MatchID        string       `json:"matchId"`
	OverRound      int          `json:"overRound"`
	ResultType     string       `json:"resultType"`
	OverReason     string       `json:"overReason"`
	WinnerPlayerID int          `json:"winnerPlayerId"`
	Players        []OverPlayer `json:"players"`
}

// OverPlayer is over.players[] (通信协议 第 9 章).
type OverPlayer struct {
	PlayerID           int         `json:"playerId"`
	PlayerName         string      `json:"playerName"`
	Camp               int         `json:"camp"`
	Online             bool        `json:"online"`
	Delivered          bool        `json:"delivered"`
	Retired            bool        `json:"retired"`
	DeliverRound       int         `json:"deliverRound"`
	Progress           float64     `json:"progress"`
	Freshness          float64     `json:"freshness"`
	GoodFruit          int         `json:"goodFruit"`
	BadFruit           int         `json:"badFruit"`
	TaskScore          int         `json:"taskScore"`
	BountyScore        int         `json:"bountyScore"`
	MainRoute          string      `json:"mainRoute"`
	RoadRounds         int         `json:"roadRounds"`
	WaterRounds        int         `json:"waterRounds"`
	MountainRounds     int         `json:"mountainRounds"`
	BranchRounds       int         `json:"branchRounds"`
	RouteSwitchCount   int         `json:"routeSwitchCount"`
	RouteTaskScore     string      `json:"routeTaskScore"`
	RouteResourceCount string      `json:"routeResourceCount"`
	TotalScore         int         `json:"totalScore"`
	PenaltyScore       int         `json:"penaltyScore"`
	ScoreDetail        ScoreDetail `json:"scoreDetail"`
	TotalGold          int         `json:"totalGold"`
}

// Result type constants.
const (
	ResultNormal  = "NORMAL"
	ResultDraw    = "DRAW"
	ResultForfeit = "FORFEIT"
	ResultInvalid = "INVALID"
)
