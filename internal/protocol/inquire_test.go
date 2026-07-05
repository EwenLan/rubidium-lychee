package protocol

import (
	"encoding/json"
	"os"
	"testing"
)

// TestParseRealInquireSample parses the real server inquire sample
// (request example.json at repo root) and verifies all key fields
// are captured correctly by the protocol structs.
func TestParseRealInquireSample(t *testing.T) {
	data, err := os.ReadFile("../../request example.json")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.MsgName != MsgInquire {
		t.Fatalf("expected msg_name=%s, got %s", MsgInquire, env.MsgName)
	}

	var inq InquireMessage
	if err := json.Unmarshal(env.MsgData, &inq); err != nil {
		t.Fatalf("unmarshal inquire: %v", err)
	}

	// Top-level scalars.
	if inq.MatchID != "match_20260704_035558_305_2710_vs_2931_744aa446" {
		t.Errorf("matchId: got %q", inq.MatchID)
	}
	if inq.Round != 2 {
		t.Errorf("round: got %d", inq.Round)
	}
	if inq.Tick != 1 {
		t.Errorf("tick: got %d", inq.Tick)
	}
	if inq.Phase != PhaseNormal {
		t.Errorf("phase: got %q", inq.Phase)
	}
	if inq.RulesVersion != "4.1" {
		t.Errorf("rulesVersion: got %q", inq.RulesVersion)
	}

	// Collections.
	if len(inq.Players) != 2 {
		t.Errorf("players: got %d, want 2", len(inq.Players))
	}
	if len(inq.Nodes) != 15 {
		t.Errorf("nodes: got %d, want 15", len(inq.Nodes))
	}
	if len(inq.Edges) != 21 {
		t.Errorf("edges: got %d, want 21", len(inq.Edges))
	}
	if len(inq.Tasks) != 3 {
		t.Errorf("tasks: got %d, want 3", len(inq.Tasks))
	}
	if len(inq.Events) != 7 {
		t.Errorf("events: got %d, want 7", len(inq.Events))
	}
	if len(inq.ActionResults) != 2 {
		t.Errorf("actionResult: got %d, want 2", len(inq.ActionResults))
	}
	if len(inq.Bounties) != 0 {
		t.Errorf("bounties: got %d, want 0", len(inq.Bounties))
	}
	if len(inq.Contests) != 0 {
		t.Errorf("contests: got %d, want 0", len(inq.Contests))
	}
	if inq.ScorePreview["RED"] != 0 || inq.ScorePreview["BLUE"] != 0 {
		t.Errorf("scorePreview: got %+v", inq.ScorePreview)
	}

	// BLUE player (2931) is MOVING on edge E01.
	var blue *PlayerState
	for i := range inq.Players {
		if inq.Players[i].PlayerID == 2931 {
			blue = &inq.Players[i]
		}
	}
	if blue == nil {
		t.Fatalf("player 2931 not found")
	}
	if blue.TeamID != "BLUE" {
		t.Errorf("blue teamId: got %q", blue.TeamID)
	}
	if blue.State != StateMoving {
		t.Errorf("blue state: got %q, want MOVING", blue.State)
	}
	if blue.CurrentNodeID != "S01" {
		t.Errorf("blue currentNodeId: got %q", blue.CurrentNodeID)
	}
	if blue.NextNodeID != "S02" {
		t.Errorf("blue nextNodeId: got %q", blue.NextNodeID)
	}
	if blue.RouteEdgeID != "E01" {
		t.Errorf("blue routeEdgeId: got %q", blue.RouteEdgeID)
	}
	if blue.RouteType != "ROAD" {
		t.Errorf("blue routeType: got %q", blue.RouteType)
	}
	if blue.MoveDirection != "FORWARD" {
		t.Errorf("blue moveDirection: got %q", blue.MoveDirection)
	}
	if blue.EdgeTotalMs != 41400 {
		t.Errorf("blue edgeTotalMs: got %d, want 41400", blue.EdgeTotalMs)
	}
	if blue.EdgeProgressMs != 1000 {
		t.Errorf("blue edgeProgressMs: got %d, want 1000", blue.EdgeProgressMs)
	}
	if blue.Freshness != 99.95 {
		t.Errorf("blue freshness: got %v, want 99.95", blue.Freshness)
	}
	if blue.GoodFruit != 100 {
		t.Errorf("blue goodFruit: got %d", blue.GoodFruit)
	}
	if blue.SquadAvailable != 8 {
		t.Errorf("blue squadAvailable: got %d", blue.SquadAvailable)
	}
	if blue.GuardActionPoint != 4 {
		t.Errorf("blue guardActionPoint: got %d", blue.GuardActionPoint)
	}

	// RED player (2710) is IDLE at S01.
	var red *PlayerState
	for i := range inq.Players {
		if inq.Players[i].PlayerID == 2710 {
			red = &inq.Players[i]
		}
	}
	if red == nil {
		t.Fatalf("player 2710 not found")
	}
	if red.State != StateIdle {
		t.Errorf("red state: got %q, want IDLE", red.State)
	}
	if red.TeamID != "RED" {
		t.Errorf("red teamId: got %q", red.TeamID)
	}

	// S08 has obstacle.
	var s08 *NodeState
	for i := range inq.Nodes {
		if inq.Nodes[i].NodeID == "S08" {
			s08 = &inq.Nodes[i]
		}
	}
	if s08 == nil {
		t.Fatalf("node S08 not found")
	}
	if !s08.HasObstacle {
		t.Errorf("S08 hasObstacle: got false, want true")
	}
	if s08.ObstacleType != "ROCKFALL" {
		t.Errorf("S08 obstacleType: got %q", s08.ObstacleType)
	}

	// S03 has resources.
	var s03 *NodeState
	for i := range inq.Nodes {
		if inq.Nodes[i].NodeID == "S03" {
			s03 = &inq.Nodes[i]
		}
	}
	if s03 == nil {
		t.Fatalf("node S03 not found")
	}
	if s03.ResourceStock["ICE_BOX"] != 1 || s03.ResourceStock["PASS_TOKEN"] != 1 || s03.ResourceStock["INTEL"] != 1 {
		t.Errorf("S03 resourceStock: got %+v", s03.ResourceStock)
	}

	// S14 is the gate with VERIFY process.
	var s14 *NodeState
	for i := range inq.Nodes {
		if inq.Nodes[i].NodeID == "S14" {
			s14 = &inq.Nodes[i]
		}
	}
	if s14 == nil {
		t.Fatalf("node S14 not found")
	}
	if s14.ProcessType != "VERIFY" {
		t.Errorf("S14 processType: got %q", s14.ProcessType)
	}
	if s14.ProcessRound != 6 {
		t.Errorf("S14 processRound: got %d", s14.ProcessRound)
	}

	// Tasks.
	var t001 *Task
	for i := range inq.Tasks {
		if inq.Tasks[i].TaskID == "T_001" {
			t001 = &inq.Tasks[i]
		}
	}
	if t001 == nil {
		t.Fatalf("task T_001 not found")
	}
	if t001.TaskTemplateID != "T01" {
		t.Errorf("T_001 taskTemplateId: got %q", t001.TaskTemplateID)
	}
	if t001.NodeID != "S03" {
		t.Errorf("T_001 nodeId: got %q", t001.NodeID)
	}
	if t001.ExpireRound != 221 {
		t.Errorf("T_001 expireRound: got %d", t001.ExpireRound)
	}
	if t001.Score != 30 {
		t.Errorf("T_001 score: got %d", t001.Score)
	}

	// Action results: first entry rejected, second accepted.
	if inq.ActionResults[0].PlayerID != 2710 {
		t.Errorf("actionResult[0] playerId: got %d", inq.ActionResults[0].PlayerID)
	}
	if inq.ActionResults[0].Result != "ACTION_REJECTED" {
		t.Errorf("actionResult[0] result: got %q", inq.ActionResults[0].Result)
	}
	if inq.ActionResults[0].ErrorCode != "TARGET_NOT_REACHABLE" {
		t.Errorf("actionResult[0] errorCode: got %q", inq.ActionResults[0].ErrorCode)
	}
	if inq.ActionResults[1].PlayerID != 2931 {
		t.Errorf("actionResult[1] playerId: got %d", inq.ActionResults[1].PlayerID)
	}
	if inq.ActionResults[1].Result != "ACCEPTED" {
		t.Errorf("actionResult[1] result: got %q", inq.ActionResults[1].Result)
	}

	// Events: check MOVE_PROGRESS for BLUE.
	var moveProg *Event
	for i := range inq.Events {
		if inq.Events[i].Type == "MOVE_PROGRESS" {
			moveProg = &inq.Events[i]
		}
	}
	if moveProg == nil {
		t.Fatalf("MOVE_PROGRESS event not found")
	}
	var mp struct {
		PlayerID        int     `json:"playerId"`
		FromNodeID      string  `json:"fromNodeId"`
		ToNodeID        string  `json:"toNodeId"`
		RouteEdgeID     string  `json:"routeEdgeId"`
		Progress        float64 `json:"progress"`
		EdgeProgressMs  int     `json:"edgeProgressMs"`
		EdgeTotalMs     int     `json:"edgeTotalMs"`
	}
	if err := json.Unmarshal(moveProg.Payload, &mp); err != nil {
		t.Fatalf("unmarshal MOVE_PROGRESS payload: %v", err)
	}
	if mp.PlayerID != 2913 {
		t.Errorf("MOVE_PROGRESS playerId: got %d, want 2913", mp.PlayerID)
	}
	if mp.FromNodeID != "S01" || mp.ToNodeID != "S02" {
		t.Errorf("MOVE_PROGRESS from/to: got %s→%s", mp.FromNodeID, mp.ToNodeID)
	}
	if mp.EdgeTotalMs != 41400 {
		t.Errorf("MOVE_PROGRESS edgeTotalMs: got %d, want 41400", mp.EdgeTotalMs)
	}
}
