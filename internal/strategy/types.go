package strategy

// PathPlan is the result of adaptive path planning.
type PathPlan struct {
	Nodes    []string  // node sequence (including start and target)
	Distance int       // total edge distance
	Blockers []Blocker // blockers on the path (empty if clear path found)
}

// Blocker represents an obstacle or enemy guard on the path.
type Blocker struct {
	NodeID  string
	Type    string // "obstacle" / "enemy_guard"
	Defense int    // guard defense (0 for obstacles)
}

// ArrivalPrediction predicts arrival frames at nodes on the planned path.
type ArrivalPrediction struct {
	NodeArrival   map[string]int // nodeID → estimated arrival frame
	GateFrame     int            // arrival at gate (0 if not on path)
	TerminalFrame int            // arrival at terminal (0 if not on path)
}

// SquadDispatch tracks an in-flight squad action.
type SquadDispatch struct {
	SubmitFrame  int
	ArrivalFrame int
	Action       string
	TargetNodeID string
	Cost         int
}

// SquadIntent represents a candidate squad action with priority.
type SquadIntent struct {
	Name     string
	Priority int
	Action   string
	Target   string
}
