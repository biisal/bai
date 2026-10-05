package notifier

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	herdrSource    = "custom:bai"
	herdrAgentName = "bai"

	// States reported to Herdr via reportState.
	stateIdle    = "idle"
	stateWorking = "working"
	stateBlocked = "blocked"

	// herdrCallTimeout bounds every herdr invocation so a hung server can
	// never slow bai down.
	herdrCallTimeout = 2 * time.Second
)

// herdrClient reports bai's state to a running Herdr server by shelling out
// to the herdr binary. A nil *herdrClient means "not inside Herdr"; every
// method is then a no-op.
type herdrClient struct {
	paneID string
	bin    string
	seq    atomic.Int64
}

// newHerdrClient returns nil when bai is not running inside a Herdr pane.
func newHerdrClient() *herdrClient {
	if os.Getenv("HERDR_ENV") != "1" {
		slog.Info("HERDR_ENV not set, skipping herdr")
		return nil
	}
	paneID := os.Getenv("HERDR_PANE_ID")
	bin := os.Getenv("HERDR_BIN_PATH")
	if paneID == "" || bin == "" {
		slog.Info("HERDR_PANE_ID or HERDR_BIN_PATH not set, skipping herdr", "paneID", paneID, "bin", bin)
		return nil
	}

	c := &herdrClient{paneID: paneID, bin: bin}
	// Herdr ignores any report whose seq is not higher than the last one it
	// accepted from this source, and that includes reports from earlier runs.
	// Seeding from the clock keeps seq increasing across restarts of bai.
	c.seq.Store(time.Now().UnixMilli())

	slog.Info("herdr initialized", "paneID", paneID, "bin", bin)
	return c
}

// reportState sends a state update. state is stateIdle, stateWorking or stateBlocked.
func (c *herdrClient) reportState(ctx context.Context, state, message string) {
	if c == nil {
		return
	}
	args := c.reportArgs(state)
	if message != "" {
		args = append(args, "--message", message)
	}
	go c.run(ctx, args)
}

// reportSession sends the resume command used for session restore.
// sessionID is an opaque identifier the agent uses to restore its session.
func (c *herdrClient) reportSession(ctx context.Context, sessionID string, resumeArgs []string) {
	if c == nil {
		return
	}
	args := c.reportArgs(stateIdle)
	if sessionID != "" {
		args = append(args, "--agent-session-id", sessionID)
	}
	if len(resumeArgs) > 0 {
		args = append(args, "--")
		args = append(args, resumeArgs...)
	}
	go c.run(ctx, args)
}

// release clears bai's agent state from Herdr. It runs synchronously (bounded
// by herdrCallTimeout) because the process is about to exit and a
// fire-and-forget call could be lost.
func (c *herdrClient) release(ctx context.Context) {
	if c == nil {
		return
	}
	c.run(ctx, []string{
		"pane", "release-agent", c.paneID,
		"--source", herdrSource,
		"--agent", herdrAgentName,
		"--seq", c.nextSeq(),
	})
}

func (c *herdrClient) reportArgs(state string) []string {
	return []string{
		"pane", "report-agent", c.paneID,
		"--source", herdrSource,
		"--agent", herdrAgentName,
		"--state", state,
		"--seq", c.nextSeq(),
	}
}

// nextSeq guarantees late or out-of-order reports cannot overwrite newer state.
func (c *herdrClient) nextSeq() string {
	return strconv.FormatInt(c.seq.Add(1), 10)
}

func (c *herdrClient) run(ctx context.Context, args []string) {
	if ctx == nil {
		ctx = context.Background()
	}
	// Detach from the caller's cancellation: a cancelled ctx (Ctrl+C, finished
	// turn) must not kill an in-flight state report.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), herdrCallTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, c.bin, args...).CombinedOutput()
	if err != nil {
		// Warn, not debug: herdr's own error (e.g. resume_not_accepted,
		// invalid_resume_argv) is in out and would otherwise be invisible.
		slog.Warn("herdr report failed", "args", args, "error", err, "output", string(out))
		return
	}
	slog.Debug("herdr report ok", "args", args)
}
