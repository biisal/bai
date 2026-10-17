package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/tui/styles"
)

const notificationTTL = 2 * time.Second

// NotifyMsg shows a short-lived notification at the top-right of the screen.
type NotifyMsg string

// Notify returns a command that shows text as a notification.
func Notify(text string) tea.Cmd {
	return func() tea.Msg { return NotifyMsg(text) }
}

// notifyExpiredMsg hides the notification with that id, unless a newer one replaced it.
type notifyExpiredMsg int

// showNotice sets the notification and returns the command that expires it.
func (c *Component) showNotice(text string) tea.Cmd {
	c.noticeID++
	c.notice = text
	id := c.noticeID
	return tea.Tick(notificationTTL, func(time.Time) tea.Msg { return notifyExpiredMsg(id) })
}

func (c *Component) expireNotice(id int) {
	if id == c.noticeID {
		c.notice = ""
	}
}

// noticeView returns the rendered notification, or "" when there is none.
func (c *Component) noticeView() string {
	if c.notice == "" {
		return ""
	}
	return styles.StyleNotification.Render(c.notice)
}
