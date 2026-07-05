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
	target := gameMap.TerminalID
	if !self.Verified {
		target = gameMap.GateID
	}
	s.lastPlan = s.planPath(state, gameMap, self.CurrentNodeID, target)
	s.lastPred = s.predictArrivals(state, gameMap, s.lastPlan)

	// Update in-flight squads (remove landed).
	s.updateInFlight(state.Round)

	if s.lastPlan != nil && len(s.lastPlan.Blockers) > 0 {
		log.Debugf("round %d: path to %s has %d blockers", state.Round, target, len(s.lastPlan.Blockers))
	}

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

// decideMainCart returns the main-cart action for this frame, or nil.
func (s *Strategy) decideMainCart(state *game.State, gameMap *game.GameMap) *protocol.Action {
	self := state.Self

	// Delivered or not IDLE: no main cart action.
	if self.Delivered || self.State != protocol.StateIdle {
		return nil
	}

	// At S14 + RUSH + not verified: verify.
	if self.CurrentNodeID == gameMap.GateID &&
		state.Phase == protocol.PhaseRush &&
		!self.Verified {
		return &protocol.Action{Action: protocol.ActionVerifyGate, TargetNodeID: gameMap.GateID}
	}

	// At S15 + verified: deliver.
	if self.CurrentNodeID == gameMap.TerminalID &&
		self.Verified &&
		self.GoodFruit > 0 &&
		self.Freshness > 0 {
		return &protocol.Action{Action: protocol.ActionDeliver}
	}

	// At process node (non-VERIFY) + not yet processed: process.
	node := gameMap.Node(self.CurrentNodeID)
	if node != nil &&
		node.ProcessType != "" &&
		node.ProcessType != "VERIFY" &&
		!s.processedAt[self.CurrentNodeID] {
		return &protocol.Action{Action: protocol.ActionProcess, TargetNodeID: self.CurrentNodeID}
	}

	// MOVE along the planned path. If next step is blocked, wait for squad
	// to clear/weaken it (squad intent handles this).
	if s.lastPlan == nil || len(s.lastPlan.Nodes) < 2 {
		return nil
	}
	nextNode := s.lastPlan.Nodes[1]
	for _, b := range s.lastPlan.Blockers {
		if b.NodeID == nextNode {
			log.Debugf("round %d: next node %s blocked (%s), waiting for squad", state.Round, nextNode, b.Type)
			return nil
		}
	}
	return &protocol.Action{Action: protocol.ActionMove, TargetNodeID: nextNode}
}
