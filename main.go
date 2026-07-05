package main

import (
	"encoding/json"
	"flag"
	"strconv"
	"time"

	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/log"
	"rubidium-lychee/internal/protocol"
	"rubidium-lychee/internal/strategy"
	"rubidium-lychee/internal/transport"
)

func main() {
	playerIDStr := flag.String("playerId", "", "player id assigned by platform (numeric)")
	host := flag.String("host", "", "match server host")
	port := flag.String("port", "", "match server port")
	playerName := flag.String("playerName", "rubidium-lychee", "player name for registration")
	version := flag.String("version", "0.1.0", "client version")
	verbose := flag.Bool("verbose", true, "log per-frame state")
	logDir := flag.String("log-dir", "logs", "log directory")
	logFile := flag.String("log-file", "client.log", "log file name")
	flag.Parse()

	if err := log.Init(*logDir, *logFile); err != nil {
		log.Fatalf("init log: %v", err)
	}
	defer log.Close()

	if *playerIDStr == "" || *host == "" || *port == "" {
		log.Fatalf("missing required args: playerId=%q host=%q port=%q", *playerIDStr, *host, *port)
	}
	playerID, err := strconv.Atoi(*playerIDStr)
	if err != nil {
		log.Fatalf("playerId must be numeric, got %q: %v", *playerIDStr, err)
	}
	log.Infof("client starting: playerId=%d host=%s port=%s", playerID, *host, *port)

	conn, err := transport.Dial(*host, *port)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	log.Infof("connected to %s:%s", *host, *port)

	// 1. Send registration.
	reg := protocol.Registration{
		PlayerID:   playerID,
		PlayerName: *playerName,
		Version:    *version,
	}
	if err := protocol.Send(conn, protocol.MsgRegistration, reg); err != nil {
		log.Fatalf("send registration: %v", err)
	}

	// 2. Receive start.
	env, err := protocol.Recv(conn)
	if err != nil {
		log.Fatalf("recv start: %v", err)
	}
	if env.MsgName != protocol.MsgStart {
		log.Fatalf("expected start, got %s", env.MsgName)
	}
	var start protocol.StartMessage
	if err := json.Unmarshal(env.MsgData, &start); err != nil {
		log.Errorf("unmarshal start (partial parse, continuing): %v", err)
	}
	log.Infof("match %s started; duration=%d rounds; %d nodes, %d edges",
		start.MatchID, start.DurationRound, len(start.Nodes), len(start.Edges))

	gameMap, err := game.BuildMap(&start)
	if err != nil {
		log.Fatalf("build map: %v", err)
	}
	log.Infof("map: start=%s gate=%s terminal=%s", gameMap.StartID, gameMap.GateID, gameMap.TerminalID)

	// 3. Send ready.
	ready := protocol.Ready{
		MatchID:  start.MatchID,
		Round:    start.Round,
		PlayerID: playerID,
	}
	if err := protocol.Send(conn, protocol.MsgReady, ready); err != nil {
		log.Fatalf("send ready: %v", err)
	}

	strat := strategy.New()

	// 4. Inquire/action loop.
	for {
		conn.SetDeadline(time.Now().Add(60 * time.Second))
		env, err := protocol.Recv(conn)
		if err != nil {
			log.Fatalf("recv: %v", err)
		}
		switch env.MsgName {
		case protocol.MsgInquire:
			var inq protocol.InquireMessage
			if err := json.Unmarshal(env.MsgData, &inq); err != nil {
				log.Errorf("unmarshal inquire: %v", err)
				continue
			}
			state, err := game.BuildState(&inq, playerID)
			if err != nil {
				log.Errorf("build state: %v", err)
				continue
			}
			actions := strat.Decide(state, gameMap)
			if *verbose {
				self := state.Self
				log.Infof("round %d phase=%s: at %s state=%s verified=%v delivered=%v → %d actions",
					state.Round, state.Phase, self.CurrentNodeID, self.State, self.Verified, self.Delivered, len(actions))
			}
			act := protocol.ActionMessage{
				MatchID:  inq.MatchID,
				Round:    inq.Round,
				PlayerID: playerID,
				Actions:  actions,
			}
			if err := protocol.Send(conn, protocol.MsgAction, act); err != nil {
				log.Fatalf("send action: %v", err)
			}
		case protocol.MsgOver:
			var over protocol.OverMessage
			if err := json.Unmarshal(env.MsgData, &over); err != nil {
				log.Errorf("unmarshal over: %v", err)
			}
			log.Infof("match over: result=%s winner=%d overRound=%d reason=%s",
				over.ResultType, over.WinnerPlayerID, over.OverRound, over.OverReason)
			for _, p := range over.Players {
				log.Infof("  player %d: delivered=%v totalScore=%d freshness=%.1f goodFruit=%d",
					p.PlayerID, p.Delivered, p.TotalScore, p.Freshness, p.GoodFruit)
			}
			return
		case protocol.MsgError:
			var errMsg protocol.ErrorMessage
			if err := json.Unmarshal(env.MsgData, &errMsg); err != nil {
				log.Errorf("unmarshal error: %v", err)
			}
			log.Errorf("server error: code=%s msg=%s round=%d", errMsg.ErrorCode, errMsg.Message, errMsg.Round)
			continue
		default:
			log.Warnf("unexpected msg_name: %s", env.MsgName)
		}
	}
}
