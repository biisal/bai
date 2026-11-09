package notifier

import (
	"fmt"
	"os"
)

const (
	herdrAgentName = "bai"

	stateIdle    = "idle"
	stateWorking = "working"
	stateBlocked = "blocked"
)

type Notifier struct {
	tty *os.File
}

func New() *Notifier {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return &Notifier{}
	}
	return &Notifier{tty: tty}
}

func (n *Notifier) report(state string) {
	if n == nil || n.tty == nil {
		return
	}
	_, _ = fmt.Fprintf(n.tty, "\x1b]7501;state=%s:app=%s\x1b\\", state, herdrAgentName)
}

func (n *Notifier) SessionStarted() {
	n.report(stateIdle)
}

func (n *Notifier) Working() {
	n.report(stateWorking)
}

func (n *Notifier) Done() {
	n.report(stateIdle)
}

func (n *Notifier) Failed() {
	n.report(stateIdle)
}

func (n *Notifier) Blocked() {
	n.report(stateBlocked)
}

func (n *Notifier) Close() {
	if n == nil || n.tty == nil {
		return
	}
	_ = n.tty.Close()
}
