// Command mockserver is a local server for integration testing.
//
// Live mode (default): drives a simplified match (MOVE/PROCESS/VERIFY/DELIVER)
// against a single client, sending over when the client delivers or the round
// limit is reached.
//
// Replay mode (--mode replay --replay <logfile>): reads a client log file,
// extracts all server-sent (RECV) messages, and replays them verbatim to a
// connecting client. Client actions are received as timing signals but their
// content is ignored. This lets you retest strategy changes against a recorded
// match.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/log"
	"rubidium-lychee/internal/protocol"
	"rubidium-lychee/internal/transport"
)

func main() {
	mode := flag.String("mode", "live", "live or replay")
	port := flag.String("port", "8081", "listen port")
	// Live mode flags.
	maxRounds := flag.Int("rounds", 600, "max rounds before forced over (live mode)")
	rushFrame := flag.Int("rush-frame", 390, "frame at which RUSH phase can trigger (live mode)")
	speed := flag.Int("speed", 10, "distance units advanced per frame (live mode)")
	obstacles := flag.String("obstacles", "", "obstacles to place at start, e.g. 'S08:ROCKFALL,S10:MUD' (live mode)")
	enemyGuards := flag.String("enemy-guards", "", "enemy guards to place, e.g. 'S10:3,S11:2' (live mode)")
	// Replay mode flags.
	replayFile := flag.String("replay", "", "client log file to replay (replay mode)")
	// Logging flags.
	logDir := flag.String("log-dir", "logs", "log directory (live mode)")
	logFile := flag.String("log-file", "mockserver.log", "log file name (live mode)")
	flag.Parse()

	switch *mode {
	case "live":
		runLive(*port, *maxRounds, *rushFrame, *speed, *obstacles, *enemyGuards, *logDir, *logFile)
	case "replay":
		runReplay(*port, *replayFile)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s (use live or replay)\n", *mode)
		os.Exit(1)
	}
}

func runLive(port string, maxRounds, rushFrame, speed int, obstacles, enemyGuards, logDir, logFile string) {
	if err := log.Init(logDir, logFile); err != nil {
		log.Fatalf("init log: %v", err)
	}
	defer log.Close()

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Infof("mock server (live) listening on :%s (rounds=%d rush-frame=%d speed=%d)",
		port, maxRounds, rushFrame, speed)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()
	fc := transport.NewConn(conn)

	// 1. Receive registration.
	env, err := protocol.Recv(fc)
	if err != nil {
		log.Fatalf("recv registration: %v", err)
	}
	if env.MsgName != protocol.MsgRegistration {
		log.Fatalf("expected registration, got %s", env.MsgName)
	}
	var reg protocol.Registration
	if err := json.Unmarshal(env.MsgData, &reg); err != nil {
		log.Fatalf("unmarshal registration: %v", err)
	}
	log.Infof("client registered: playerId=%d name=%s version=%s", reg.PlayerID, reg.PlayerName, reg.Version)

	// 2. Send start.
	start := buildStart(reg.PlayerID, reg.PlayerName)
	if err := protocol.Send(fc, protocol.MsgStart, start); err != nil {
		log.Fatalf("send start: %v", err)
	}

	// 3. Receive ready.
	env, err = protocol.Recv(fc)
	if err != nil {
		log.Fatalf("recv ready: %v", err)
	}
	if env.MsgName != protocol.MsgReady {
		log.Fatalf("expected ready, got %s", env.MsgName)
	}
	log.Infof("client ready, starting frame loop")

	// 4. Build game map and engine.
	gameMap, err := game.BuildMap(&start)
	if err != nil {
		log.Fatalf("build map: %v", err)
	}
	eng := newEngine(start.MatchID, reg.PlayerID, gameMap, rushFrame, speed)
	eng.playerName = reg.PlayerName
	eng.generateObstacles(parseObstacles(obstacles))
	for nodeID, def := range parseEnemyGuards(enemyGuards) {
		eng.placeEnemyGuard(nodeID, def)
	}

	// 5. Frame loop.
	for {
		eng.nextRound()
		eng.preFrame()
		inq := eng.buildInquire()
		if err := protocol.Send(fc, protocol.MsgInquire, inq); err != nil {
			log.Fatalf("send inquire %d: %v", eng.round, err)
		}

		fc.SetDeadline(time.Now().Add(30 * time.Second))
		env, err := protocol.Recv(fc)
		if err != nil {
			log.Fatalf("recv action %d: %v", eng.round, err)
		}
		if env.MsgName != protocol.MsgAction {
			log.Warnf("round %d: expected action, got %s", eng.round, env.MsgName)
			continue
		}
		var act protocol.ActionMessage
		if err := json.Unmarshal(env.MsgData, &act); err != nil {
			log.Errorf("round %d: unmarshal action: %v", eng.round, err)
		}

		if len(act.Actions) > 0 {
			eng.applyActions(act.Actions)
		}
		eng.tick()

		if eng.isOver(maxRounds) {
			break
		}
	}

	// 6. Send over.
	over := protocol.OverMessage{
		MatchID:        start.MatchID,
		OverRound:      eng.round,
		ResultType:     protocol.ResultNormal,
		OverReason:     "ALL_DELIVERED",
		WinnerPlayerID: reg.PlayerID,
		Players: []protocol.OverPlayer{
			{
				PlayerID:     reg.PlayerID,
				PlayerName:   reg.PlayerName,
				Camp:         0,
				Online:       true,
				Delivered:    eng.delivered,
				DeliverRound: func() int { if eng.delivered { return eng.round }; return 0 }(),
				Freshness:    100.0,
				GoodFruit:    100,
				TotalScore:   scoreFor(eng.delivered),
				ScoreDetail:  protocol.ScoreDetail{Delivery: scoreFor(eng.delivered), Total: scoreFor(eng.delivered)},
			},
			{
				PlayerID:    9999,
				PlayerName:  "mock-opponent",
				Camp:        1,
				Online:      true,
				Freshness:   100.0,
				GoodFruit:   100,
				ScoreDetail: protocol.ScoreDetail{},
			},
		},
	}
	if !eng.delivered {
		over.OverReason = "TIME_LIMIT"
		over.WinnerPlayerID = 0
	}
	if err := protocol.Send(fc, protocol.MsgOver, over); err != nil {
		log.Fatalf("send over: %v", err)
	}
	log.Infof("match over: round=%d delivered=%v", eng.round, eng.delivered)
	time.Sleep(100 * time.Millisecond)
}

// runReplay reads a client log file and replays all server-sent (RECV)
// messages verbatim to a connecting client. Client actions are received as
// timing signals but their content does not affect the replayed frames.
func runReplay(port, replayFile string) {
	log.InitStdout()
	defer log.Close()

	if replayFile == "" {
		log.Fatalf("replay mode requires --replay <logfile>")
	}
	log.Infof("parsing log: %s", replayFile)
	msgs, err := log.ParseLog(replayFile)
	if err != nil {
		log.Fatalf("parse log: %v", err)
	}
	recvMsgs := log.FilterRecv(msgs)
	sendCount := len(msgs) - len(recvMsgs)
	log.Infof("found %d server messages (RECV), %d client messages (SEND)",
		len(recvMsgs), sendCount)
	if len(recvMsgs) == 0 {
		log.Fatalf("no server messages found in log")
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	log.Infof("mock server (replay) listening on :%s", port)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()
	fc := transport.NewConn(conn)

	// Receive registration first (not part of the recorded RECV stream).
	fc.SetDeadline(time.Now().Add(30 * time.Second))
	env, err := protocol.Recv(fc)
	if err != nil {
		log.Fatalf("recv registration: %v", err)
	}
	if env.MsgName != protocol.MsgRegistration {
		log.Fatalf("expected registration, got %s", env.MsgName)
	}
	log.Infof("client registered, starting replay")

	// Replay each recorded server message in order. Between start and inquire
	// frames, wait for the client's response (ready / action) as a timing
	// signal; the response content is ignored.
	for _, msg := range recvMsgs {
		if err := protocol.SendRaw(fc, msg.Envelope); err != nil {
			log.Fatalf("send %s: %v", msg.MsgName, err)
		}
		log.Infof("sent %s round=%d", msg.MsgName, msg.Round)

		switch msg.MsgName {
		case protocol.MsgStart:
			fc.SetDeadline(time.Now().Add(30 * time.Second))
			if _, err := protocol.Recv(fc); err != nil {
				log.Fatalf("recv ready: %v", err)
			}
		case protocol.MsgInquire:
			fc.SetDeadline(time.Now().Add(30 * time.Second))
			if _, err := protocol.Recv(fc); err != nil {
				log.Fatalf("recv action: %v", err)
			}
		case protocol.MsgOver:
			log.Infof("replay complete")
			return
		}
	}
	log.Infof("replay ended (no over message in log)")
}

// scoreFor returns a placeholder score for the mock.
func scoreFor(delivered bool) int {
	if delivered {
		return 240 // delivery base score
	}
	return 0
}

// parseObstacles parses "S08:ROCKFALL,S10:MUD" into a map.
func parseObstacles(s string) map[string]string {
	out := make(map[string]string)
	if s == "" {
		return out
	}
	for _, part := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return out
}

// parseEnemyGuards parses "S10:3,S11:2" into a map of nodeID → defense.
func parseEnemyGuards(s string) map[string]int {
	out := make(map[string]int)
	if s == "" {
		return out
	}
	for _, part := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		def, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			continue
		}
		out[strings.TrimSpace(kv[0])] = def
	}
	return out
}
