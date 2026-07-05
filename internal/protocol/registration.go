package protocol

// Registration is sent by the client immediately after TCP connect.
type Registration struct {
	PlayerID   int    `json:"playerId"`
	PlayerName string `json:"playerName"`
	Version    string `json:"version"`
}

// Ready is sent by the client after receiving and processing start.
type Ready struct {
	MatchID  string `json:"matchId"`
	Round    int    `json:"round"`
	PlayerID int    `json:"playerId"`
}
