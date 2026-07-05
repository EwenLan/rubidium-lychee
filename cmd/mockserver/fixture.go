package main

import "rubidium-lychee/internal/protocol"

// buildStart returns a start message with the full basic map from 附录 A.
func buildStart(playerID int, playerName string) protocol.StartMessage {
	return protocol.StartMessage{
		MatchID:       "mock_001",
		RulesVersion:  "4.1",
		Round:         1,
		DurationRound: 600,
		Map: protocol.MapConfig{
			MapID:   "litchi_map_medium_a",
			MapName: "一骑红尘：荔枝争运战竞技地图",
			MaxX:    80,
			MaxY:    60,
			Gameplay: protocol.Gameplay{
				Roles: protocol.Roles{
					StartNodeID:     "S01",
					TerminalNodeIDs: []string{"S15"},
					GateNodeID:      "S14",
					SafeZoneNodeIDs: []string{"S15"},
				},
				ObstacleCandidateNodeIDs: []string{"S06", "S08", "S10", "S11"},
				ProcessNodes: []protocol.ProcessNodeDef{
					{NodeID: "S02", ProcessType: "TRANSFER", ProcessRound: 4, CanWindow: true},
					{NodeID: "S04", ProcessType: "BOARD", ProcessRound: 7, CanWindow: true},
					{NodeID: "S05", ProcessType: "WATER_TRANSFER", ProcessRound: 6, CanWindow: true},
					{NodeID: "S11", ProcessType: "PASS_TRANSFER", ProcessRound: 5, CanWindow: true},
					{NodeID: "S13", ProcessType: "PALACE_TRANSFER", ProcessRound: 5, CanWindow: true},
					{NodeID: "S14", ProcessType: "VERIFY", ProcessRound: 6, CanWindow: true},
				},
			},
		},
		Players: []protocol.PlayerInfo{
			{PlayerID: playerID, Camp: 0, TeamID: "RED", Name: playerName},
			{PlayerID: 9999, Camp: 1, TeamID: "BLUE", Name: "mock-opponent"},
		},
		Nodes:         basicMapNodes(),
		Edges:         basicMapEdges(),
		Resources:     basicMapResources(),
		TaskTemplates: basicMapTaskTemplates(),
	}
}

// basicMapResources returns the static resource placement (附录 A).
func basicMapResources() []protocol.ResourceDef {
	return []protocol.ResourceDef{
		{NodeID: "S03", ResourceType: "ICE_BOX", Count: 1, ClaimRound: 2},
		{NodeID: "S03", ResourceType: "PASS_TOKEN", Count: 1, ClaimRound: 2},
		{NodeID: "S03", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S04", ResourceType: "SHORT_HORSE", Count: 1, ClaimRound: 2},
		{NodeID: "S04", ResourceType: "BOAT_RIGHT", Count: 1, ClaimRound: 2},
		{NodeID: "S04", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S06", ResourceType: "ICE_BOX", Count: 1, ClaimRound: 2},
		{NodeID: "S06", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S07", ResourceType: "ICE_BOX", Count: 1, ClaimRound: 2},
		{NodeID: "S07", ResourceType: "SHORT_HORSE", Count: 1, ClaimRound: 2},
		{NodeID: "S08", ResourceType: "SHORT_HORSE", Count: 1, ClaimRound: 2},
		{NodeID: "S08", ResourceType: "PASS_TOKEN", Count: 1, ClaimRound: 2},
		{NodeID: "S08", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S09", ResourceType: "FAST_HORSE", Count: 1, ClaimRound: 2},
		{NodeID: "S09", ResourceType: "OFFICIAL_PERMIT", Count: 1, ClaimRound: 2},
		{NodeID: "S10", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S11", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
		{NodeID: "S13", ResourceType: "PASS_TOKEN", Count: 1, ClaimRound: 2},
		{NodeID: "S13", ResourceType: "OFFICIAL_PERMIT", Count: 1, ClaimRound: 2},
		{NodeID: "S13", ResourceType: "INTEL", Count: 1, ClaimRound: 2},
	}
}

// basicMapTaskTemplates returns the task template definitions (任务书 5.2).
func basicMapTaskTemplates() []protocol.TaskTemplate {
	return []protocol.TaskTemplate{
		{TaskTemplateID: "T01", Name: "限时过关", ProcessType: "PASS_NODE", ProcessRound: 3, Score: 30, CandidateNodeIDs: []string{"S03"}},
		{TaskTemplateID: "T02", Name: "抵驿催运", ProcessType: "CLAIM_TASK", ProcessRound: 4, Score: 30, CandidateNodeIDs: []string{"S07", "S10"}},
		{TaskTemplateID: "T04", Name: "清障任务", ProcessType: "CLEAR_OBSTACLE", ProcessRound: 6, Score: 30, CandidateNodeIDs: []string{"S06", "S08"}},
		{TaskTemplateID: "T06", Name: "争马换乘", ProcessType: "CLAIM_TASK", ProcessRound: 3, Score: 30, CandidateNodeIDs: []string{"S09", "S04", "S06"}},
		{TaskTemplateID: "T08", Name: "码头争船", ProcessType: "CLAIM_TASK", ProcessRound: 4, Score: 30, CandidateNodeIDs: []string{"S04", "S05"}},
		{TaskTemplateID: "T11", Name: "栈道复核", ProcessType: "CLAIM_TASK", ProcessRound: 4, Score: 30, CandidateNodeIDs: []string{"S08", "S10", "S11"}},
		{TaskTemplateID: "T12", Name: "官道关验", ProcessType: "CLAIM_TASK", ProcessRound: 5, Score: 15, CandidateNodeIDs: []string{"S11", "S13"}},
		{TaskTemplateID: "T13", Name: "水陆联运", ProcessType: "CLAIM_TASK", ProcessRound: 5, Score: 15, CandidateNodeIDs: []string{"S13", "S09", "S12"}},
		{TaskTemplateID: "T14", Name: "山口急递", ProcessType: "CLAIM_TASK", ProcessRound: 5, Score: 15, CandidateNodeIDs: []string{"S10", "S11", "S12"}},
	}
}

func basicMapNodes() []protocol.NodeDef {
	return []protocol.NodeDef{
		{NodeID: "S01", Name: "岭南果园", NodeType: "START", Start: true, X: 5, Y: 50},
		{NodeID: "S02", Name: "南岭驿", NodeType: "CHECKPOINT", X: 15, Y: 48},
		{NodeID: "S03", Name: "梅关驿", NodeType: "PASS", X: 22, Y: 45},
		{NodeID: "S04", Name: "江南码头", NodeType: "DOCK", X: 20, Y: 38},
		{NodeID: "S05", Name: "洞庭水驿", NodeType: "WATER_STATION", X: 30, Y: 35},
		{NodeID: "S06", Name: "五岭山道", NodeType: "MOUNTAIN_NODE", X: 15, Y: 30},
		{NodeID: "S07", Name: "荆襄大驿", NodeType: "STATION", X: 35, Y: 40},
		{NodeID: "S08", Name: "秦岭栈道", NodeType: "MOUNTAIN_PASS", X: 25, Y: 25},
		{NodeID: "S09", Name: "洛阳驿", NodeType: "STATION", X: 45, Y: 38},
		{NodeID: "S10", Name: "武关", NodeType: "KEY_PASS", X: 55, Y: 36},
		{NodeID: "S11", Name: "潼关驿", NodeType: "PASS", X: 60, Y: 33},
		{NodeID: "S12", Name: "关中平原", NodeType: "JUNCTION", X: 65, Y: 30},
		{NodeID: "S13", Name: "灞桥驿", NodeType: "PALACE_STATION", X: 70, Y: 25},
		{NodeID: "S14", Name: "朱雀门", NodeType: "GATE", X: 76, Y: 18},
		{NodeID: "S15", Name: "兴庆宫", NodeType: "FINISH", Terminal: true, X: 78, Y: 18},
	}
}

func basicMapEdges() []protocol.EdgeDef {
	return []protocol.EdgeDef{
		{EdgeID: "E01", FromNode: "S01", ToNode: "S02", FromNodeID: "S01", ToNodeID: "S02", PathID: "P_E01", RouteType: "ROAD", Distance: 30, Bidirectional: true},
		{EdgeID: "E02", FromNode: "S02", ToNode: "S03", FromNodeID: "S02", ToNodeID: "S03", PathID: "P_E02", RouteType: "ROAD", Distance: 25, Bidirectional: true},
		{EdgeID: "E03", FromNode: "S03", ToNode: "S07", FromNodeID: "S03", ToNodeID: "S07", PathID: "P_E03", RouteType: "ROAD", Distance: 54, Bidirectional: true},
		{EdgeID: "E04", FromNode: "S07", ToNode: "S09", FromNodeID: "S07", ToNodeID: "S09", PathID: "P_E04", RouteType: "ROAD", Distance: 46, Bidirectional: true},
		{EdgeID: "E05", FromNode: "S09", ToNode: "S10", FromNodeID: "S09", ToNodeID: "S10", PathID: "P_E05", RouteType: "ROAD", Distance: 40, Bidirectional: true},
		{EdgeID: "E06", FromNode: "S10", ToNode: "S11", FromNodeID: "S10", ToNodeID: "S11", PathID: "P_E06", RouteType: "ROAD", Distance: 36, Bidirectional: true},
		{EdgeID: "E07", FromNode: "S11", ToNode: "S12", FromNodeID: "S11", ToNodeID: "S12", PathID: "P_E07", RouteType: "ROAD", Distance: 20, Bidirectional: true},
		{EdgeID: "E08", FromNode: "S12", ToNode: "S13", FromNodeID: "S12", ToNodeID: "S13", PathID: "P_E08", RouteType: "ROAD", Distance: 25, Bidirectional: true},
		{EdgeID: "E09", FromNode: "S13", ToNode: "S14", FromNodeID: "S13", ToNodeID: "S14", PathID: "P_E09", RouteType: "ROAD", Distance: 18, Bidirectional: true},
		{EdgeID: "E10", FromNode: "S14", ToNode: "S15", FromNodeID: "S14", ToNodeID: "S15", PathID: "P_E10", RouteType: "ROAD", Distance: 10, Bidirectional: true},
		{EdgeID: "E11", FromNode: "S02", ToNode: "S04", FromNodeID: "S02", ToNodeID: "S04", PathID: "P_E11", RouteType: "ROAD", Distance: 20, Bidirectional: true},
		{EdgeID: "E12", FromNode: "S04", ToNode: "S05", FromNodeID: "S04", ToNodeID: "S05", PathID: "P_E12", RouteType: "WATER", Distance: 44, Bidirectional: true},
		{EdgeID: "E13", FromNode: "S05", ToNode: "S07", FromNodeID: "S05", ToNodeID: "S07", PathID: "P_E13", RouteType: "BRANCH", Distance: 46, Bidirectional: true},
		{EdgeID: "E15", FromNode: "S01", ToNode: "S06", FromNodeID: "S01", ToNodeID: "S06", PathID: "P_E15", RouteType: "MOUNTAIN", Distance: 44, Bidirectional: true},
		{EdgeID: "E16", FromNode: "S06", ToNode: "S08", FromNodeID: "S06", ToNodeID: "S08", PathID: "P_E16", RouteType: "MOUNTAIN", Distance: 54, Bidirectional: true},
		{EdgeID: "E17", FromNode: "S08", ToNode: "S10", FromNodeID: "S08", ToNodeID: "S10", PathID: "P_E17", RouteType: "BRANCH", Distance: 46, Bidirectional: true},
		{EdgeID: "E18", FromNode: "S03", ToNode: "S06", FromNodeID: "S03", ToNodeID: "S06", PathID: "P_E18", RouteType: "BRANCH", Distance: 38, Bidirectional: true},
		{EdgeID: "E19", FromNode: "S05", ToNode: "S09", FromNodeID: "S05", ToNodeID: "S09", PathID: "P_E19", RouteType: "WATER", Distance: 48, Bidirectional: true},
		{EdgeID: "E20", FromNode: "S07", ToNode: "S08", FromNodeID: "S07", ToNodeID: "S08", PathID: "P_E20", RouteType: "MOUNTAIN", Distance: 42, Bidirectional: true},
		{EdgeID: "E21", FromNode: "S04", ToNode: "S07", FromNodeID: "S04", ToNodeID: "S07", PathID: "P_E21", RouteType: "BRANCH", Distance: 54, Bidirectional: true},
		{EdgeID: "E22", FromNode: "S08", ToNode: "S09", FromNodeID: "S08", ToNodeID: "S09", PathID: "P_E22", RouteType: "BRANCH", Distance: 64, Bidirectional: true},
	}
}

// processNodeMap returns process type and round for each process node.
func processNodeMap() (map[string]string, map[string]int) {
	ptypes := map[string]string{
		"S02": "TRANSFER", "S04": "BOARD", "S05": "WATER_TRANSFER",
		"S11": "PASS_TRANSFER", "S13": "PALACE_TRANSFER", "S14": "VERIFY",
	}
	prounds := map[string]int{
		"S02": 4, "S04": 7, "S05": 6, "S11": 5, "S13": 5, "S14": 6,
	}
	return ptypes, prounds
}

// resourceStockMap returns the initial resource stock per node (from 附录 A).
func resourceStockMap() map[string]map[string]int {
	return map[string]map[string]int{
		"S03": {"ICE_BOX": 1, "PASS_TOKEN": 1, "INTEL": 1},
		"S04": {"SHORT_HORSE": 1, "BOAT_RIGHT": 1, "INTEL": 1},
		"S06": {"ICE_BOX": 1, "INTEL": 1},
		"S07": {"ICE_BOX": 1, "SHORT_HORSE": 1},
		"S08": {"SHORT_HORSE": 1, "PASS_TOKEN": 1, "INTEL": 1},
		"S09": {"FAST_HORSE": 1, "OFFICIAL_PERMIT": 1},
		"S10": {"INTEL": 1},
		"S11": {"INTEL": 1},
		"S13": {"PASS_TOKEN": 1, "OFFICIAL_PERMIT": 1, "INTEL": 1},
	}
}

// basicMapNodeStates returns the static base node states (without dynamic
// obstacle/guard/scout state, which the engine adds).
func basicMapNodeStates() []protocol.NodeState {
	nodes := basicMapNodes()
	ptypes, prounds := processNodeMap()
	rstocks := resourceStockMap()
	out := make([]protocol.NodeState, 0, len(nodes))
	for _, n := range nodes {
		ns := protocol.NodeState{
			NodeID:            n.NodeID,
			Name:              n.Name,
			X:                 n.X,
			Y:                 n.Y,
			NodeType:          n.NodeType,
			Start:             n.Start,
			Terminal:          n.Terminal,
			Visible:           true,
			ResourceVisible:   true,
			ResourceStock:     map[string]int{},
			Scouted:           []protocol.ScoutMarker{},
			CanWindow:         true,
			EffectiveCombatCount: 0,
			GuardBlockCount:   0,
			KeyPassCombatCount: 0,
		}
		if pt, ok := ptypes[n.NodeID]; ok {
			ns.ProcessType = pt
			ns.ProcessRound = prounds[n.NodeID]
		}
		if rs, ok := rstocks[n.NodeID]; ok {
			// Copy to avoid sharing the same map.
			ns.ResourceStock = make(map[string]int, len(rs))
			for k, v := range rs {
				ns.ResourceStock[k] = v
			}
		}
		out = append(out, ns)
	}
	return out
}
