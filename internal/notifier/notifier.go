// Package notifier tells the user what bai is doing.
//
// Inside a Herdr pane, state changes are reported to Herdr, which shows them
// in its sidebar and raises its own notifications. Outside Herdr, bai falls
// back to playing the configured notification sound.
//
// Callers only describe what happened (Working, Done, Failed, ...); how the
// user gets told is decided here.
package notifier

import (
	"context"
	"log/slog"
)

// Sound plays a notification sound. *audio.AudioPlayer satisfies it.
type Sound interface {
	Play() error
}

// Notifier is safe to use as a nil pointer: every method becomes a no-op.
type Notifier struct {
	herdr *herdrClient // nil when not running inside Herdr
	sound Sound        // optional; used only when Herdr is not available
}

// New returns a Notifier and detects whether bai is running inside Herdr.
// sound may be nil (no audio fallback). Pass a nil interface, not a typed nil
// pointer. Call it after the logger is set up.
func New(sound Sound) *Notifier {
	return &Notifier{herdr: newHerdrClient(), sound: sound}
}

// SessionStarted claims the pane in Herdr and registers how to resume bai.
// sessionID may be empty.
func (n *Notifier) SessionStarted(ctx context.Context, sessionID string) {
	if n == nil {
		return
	}
	n.herdr.reportSession(ctx, sessionID, []string{"kai", "mai"})
}

// Working marks the start of an agent turn.
func (n *Notifier) Working(ctx context.Context) {
	if n == nil {
		return
	}
	n.herdr.reportState(ctx, stateWorking, "")
}

// Done marks a successfully finished turn.
// Herdr: idle (Herdr raises its own notification). Otherwise: plays the sound.
func (n *Notifier) Done(ctx context.Context) {
	if n == nil {
		return
	}
	if n.herdr != nil {
		n.herdr.reportState(ctx, stateIdle, "")
		return
	}
	n.playSound()
}

// Failed marks a turn that ended with an error or was cancelled.
// No sound is played; Herdr just goes back to idle.
func (n *Notifier) Failed(ctx context.Context, _ error) {
	if n == nil {
		return
	}
	n.herdr.reportState(ctx, stateIdle, "")
}

// Blocked marks that the agent is waiting on the user (e.g. a permission
// prompt). Herdr shows the blocked state; no sound is played.
func (n *Notifier) Blocked(ctx context.Context) {
	if n == nil {
		return
	}
	n.herdr.reportState(ctx, stateBlocked, "")
}

// Close releases the pane in Herdr. Call once, when bai exits.
func (n *Notifier) Close(ctx context.Context) {
	if n == nil {
		return
	}
	n.herdr.release(ctx)
}

// playSound plays in the background: Play blocks until the sound finishes and
// must not hold up the caller.
func (n *Notifier) playSound() {
	if n.sound == nil {
		return
	}
	go func() {
		if err := n.sound.Play(); err != nil {
			slog.Error("failed to play notification sound", "error", err)
		}
	}()
}
