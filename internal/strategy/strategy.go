package strategy

import (
	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/log"
	"rubidium-lychee/internal/protocol"
)

// Strategy is a stateful decision maker implementing the C hybrid approach:
// intent-based skeleton (scout/clear) + reactive fill (weaken).
//
// Each frame it produces up to 2 actions: 1 main-cart + 1 squad (separate
// categories per 任务书 4.1). All decisions derive from the current inquire
// state — no hardcoded map assumptions.
type Strategy struct {
	prevNode     string
	processedAt  map[string]bool
	processingAt string

	// Adaptive state.
	selfTeamID     string
	inFlightSquads []SquadDispatch
	opponent       *OpponentModel

	// Latest plan + prediction (computed each frame, used by squad intents).
	lastPlan *PathPlan
	lastPred *ArrivalPrediction

	// MOVE rejection tracking: if we sent MOVE but cart is still IDLE next
	// frame, the MOVE was rejected (e.g. PROCESS_REQUIRED). Fall back to PROCESS.
	lastMoveTarget string
	moveRejected   bool

	// Resolved gate/terminal IDs (cached, with fallback from state.Nodes).
	gateID     string
	terminalID string
}

// New returns a fresh Strategy.
func New() *Strategy {
	return &Strategy{
		processedAt: make(map[string]bool),
		opponent:    &OpponentModel{},
	}
}

// Decide returns the actions to submit this frame. At most 1 main-cart
// action and 1 squad action (different categories, both allowed per frame).
func (s *Strategy) Decide(state *game.State, gameMap *game.GameMap) []protocol.Action {
	self := state.Self
	if self == nil {
		return nil
	}

	// Identify self team on first call.
	if s.selfTeamID == "" {
		s.selfTeamID = self.TeamID
	}

	// Resolve gate/terminal IDs (from gameMap, fallback to state.Nodes by nodeType).
	s.resolveRoleIDs(state, gameMap)

	// Detect MOVE rejection: if we sent MOVE last frame but cart is still
	// IDLE at the same node, the server rejected it (likely PROCESS_REQUIRED).
	if s.moveRejected {
		s.processedAt[self.CurrentNodeID] = false
		s.processingAt = ""
		s.moveRejected = false
		log.Infof("round %d: MOVE rejected, will try PROCESS at %s", state.Round, self.CurrentNodeID)
	}
	if s.lastMoveTarget != "" && self.State == protocol.StateIdle && self.CurrentNodeID == s.prevNode {
		s.moveRejected = true
	}
	s.lastMoveTarget = ""

	// Update opponent model every frame.
	s.updateOpponent(state, gameMap)

	// Detect arrival at a new node: reset per-visit process state.
	if self.CurrentNodeID != s.prevNode {
		s.processedAt[self.CurrentNodeID] = false
		s.processingAt = ""
		s.prevNode = self.CurrentNodeID
	}

	// Track process completion.
	if self.State == protocol.StateProcessing || self.State == protocol.StateVerifying {
		s.processingAt = self.CurrentNodeID
	} else if self.State == protocol.StateIdle && s.processingAt == self.CurrentNodeID && s.processingAt != "" {
		s.processedAt[self.CurrentNodeID] = true
		s.processingAt = ""
	}

	// Plan path + predict arrivals every frame (adaptive).
	target := s.terminalID
	if !self.Verified {
		target = s.gateID
	}
	if target == "" {
		log.Errorf("round %d: cannot resolve target (gate=%s terminal=%s)", state.Round, s.gateID, s.terminalID)
		return nil
	}
	s.lastPlan = s.planPath(state, gameMap, self.CurrentNodeID, target)
	s.lastPred = s.predictArrivals(state, gameMap, s.lastPlan)

	if s.lastPlan == nil {
		log.Warnf("round %d: no path from %s to %s", state.Round, self.CurrentNodeID, target)
	} else if len(s.lastPlan.Blockers) > 0 {
		log.Debugf("round %d: path to %s has %d blockers", state.Round, target, len(s.lastPlan.Blockers))
	}

	// Update in-flight squads (remove landed).
	s.updateInFlight(state.Round)

	// Collect actions: at most 1 main cart + 1 squad.
	var actions []protocol.Action
	if mainAct := s.decideMainCart(state, gameMap); mainAct != nil {
		actions = append(actions, *mainAct)
	}
	if squadAct := s.decideSquad(state, gameMap); squadAct != nil {
		actions = append(actions, *squadAct)
	}
	return actions
}

// resolveRoleIDs resolves gate and terminal IDs from gameMap, falling back
// to state.Nodes by nodeType if gameMap doesn't have them.
func (s *Strategy) resolveRoleIDs(state *game.State, gameMap *game.GameMap) {
	if s.gateID == "" {
		s.gateID = gameMap.GateID
	}
	if s.terminalID == "" {
		s.terminalID = gameMap.TerminalID
	}
	// Fallback: search state.Nodes by nodeType.
	if s.gateID == "" {
		for id, ns := range state.Nodes {
			if ns.NodeType == "GATE" {
				s.gateID = id
				break
			}
		}
	}
	if s.terminalID == "" {
		for id, ns := range state.Nodes {
			if ns.NodeType == "FINISH" || ns.Terminal {
				s.terminalID = id
				break
			}
		}
	}
}

// getProcessType returns the process type for a node, preferring the inquire's
// per-frame state over the static gameMap (which depends on start message parsing).
func getProcessType(state *game.State, gameMap *game.GameMap, nodeID string) string {
	if ns := state.Node(nodeID); ns != nil && ns.ProcessType != "" {
		return ns.ProcessType
	}
	if n := gameMap.Node(nodeID); n != nil {
		return n.ProcessType
	}
	return ""
}

// decideMainCart returns the main-cart action for this frame, or nil.
func (s *Strategy) decideMainCart(state *game.State, gameMap *game.GameMap) *protocol.Action {
	self := state.Self

	// Delivered or not IDLE: no main cart action.
	if self.Delivered || self.State != protocol.StateIdle {
		return nil
	}

	// At gate + RUSH + not verified: verify.
	if self.CurrentNodeID == s.gateID &&
		state.Phase == protocol.PhaseRush &&
		!self.Verified {
		return &protocol.Action{Action: protocol.ActionVerifyGate, TargetNodeID: s.gateID}
	}

	// At terminal + verified: deliver.
	if self.CurrentNodeID == s.terminalID &&
		self.Verified &&
		self.GoodFruit > 0 &&
		self.Freshness > 0 {
		return &protocol.Action{Action: protocol.ActionDeliver}
	}

	// At process node (non-VERIFY) + not yet processed: process.
	// Use inquire's processType (state.Nodes) first, fallback to gameMap.
	pt := getProcessType(state, gameMap, self.CurrentNodeID)
	if pt != "" && pt != "VERIFY" && !s.processedAt[self.CurrentNodeID] {
		return &protocol.Action{Action: protocol.ActionProcess, TargetNodeID: self.CurrentNodeID}
	}

	// MOVE along the planned path. If next step is blocked, wait for squad
	// to clear/weaken it (squad intent handles this).
	if s.lastPlan == nil || len(s.lastPlan.Nodes) < 2 {
		// No planned path — try basic shortest path as fallback.
		path, _ := gameMap.ShortestPath(self.CurrentNodeID, s.gateID)
		if !self.Verified && path != nil && len(path) >= 2 {
			next := path[1]
			s.lastMoveTarget = next
			return &protocol.Action{Action: protocol.ActionMove, TargetNodeID: next}
		}
		path, _ = gameMap.ShortestPath(self.CurrentNodeID, s.terminalID)
		if path != nil && len(path) >= 2 {
			next := path[1]
			s.lastMoveTarget = next
			return &protocol.Action{Action: protocol.ActionMove, TargetNodeID: next}
		}
		return nil
	}
	nextNode := s.lastPlan.Nodes[1]
	for _, b := range s.lastPlan.Blockers {
		if b.NodeID == nextNode {
			log.Debugf("round %d: next node %s blocked (%s), waiting for squad", state.Round, nextNode, b.Type)
			return nil
		}
	}
	s.lastMoveTarget = nextNode
	return &protocol.Action{Action: protocol.ActionMove, TargetNodeID: nextNode}
}
