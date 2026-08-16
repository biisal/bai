package commands

import (
	"context"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
)

type CommandItem struct {
	Name string
	Desc string
}

func (i CommandItem) Title() string {
	return i.Name
}

func (i CommandItem) Description() string {
	return i.Desc
}

func (i CommandItem) FilterValue() string {
	return i.Name
}

func toListItems[T list.Item](items []T) []list.Item {
	out := make([]list.Item, len(items))
	for i, it := range items {
		out[i] = it
	}
	return out
}

var rootCommand = ""

type Commands struct {
	List     list.Model
	Current  string
	ShowList bool
	Width    int
	commands map[string]func() []list.Item
	gateway  *agent.Gateway

	lastSynced string
}

func NewCommands(ctx context.Context, providers []config.ProviderConfig, gateway *agent.Gateway) *Commands {
	models := parseModels(providers)
	commands := map[string]func() []list.Item{
		"": func() []list.Item {
			return toListItems([]CommandItem{
				{
					Name: "models",
					Desc: "show available models",
				},
				{
					Name: "sessions",
					Desc: "show list of conversations",
				},
				{
					Name: "exit", // TODO : implement funcionality
					Desc: "exit the application",
				},
			})
		},
		"models": func() []list.Item {
			return models
		},
		"sessions": func() []list.Item {
			return toListItems(parseConversations(ctx, gateway.GetConversationsByCurrentDir))
		},
	}

	listStyles := newStyles(true, 0)
	list := list.New(commands[rootCommand](), itemDelegate{styles: &listStyles}, 5, 10)
	list.SetShowStatusBar(false)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowFilter(false)
	list.KeyMap.Quit.SetKeys("ctrl+c") // TODO : find a proper solution
	list.KeyMap.ForceQuit.SetKeys("ctrl+c")
	return &Commands{
		List:     list,
		Current:  rootCommand,
		commands: commands,
		gateway:  gateway,
	}
}

func (c *Commands) Update(command string) {
	if c.Current == command {
		return
	}

	fn, ok := c.commands[command]
	if !ok {
		slog.Warn("update_commands", "command", command)
		c.ShowList = false
		return
	}

	c.Current = command
	c.List.SetItems(fn())
	slog.Debug("update", "current", c.Current)
}

func (c *Commands) SetSize(width int) {
	c.Width = width
	listStyles := newStyles(true, width)
	d := itemDelegate{styles: &listStyles}
	c.List.SetDelegate(d)
}

func (c *Commands) View() string {
	if !c.ShowList {
		return ""
	}
	c.List.SetWidth(c.Width)
	slog.Debug("view", "current", c.Current)
	return c.List.View()
}

func (c *Commands) Sync(text string) {
	if !strings.HasPrefix(text, "/") {
		c.ShowList = false
		c.lastSynced = ""
		return
	}

	if c.lastSynced == text {
		return
	}
	c.lastSynced = text

	text = text[1:]
	if cmd, filter, found := strings.Cut(text, " "); found {
		if _, ok := c.commands[cmd]; !ok {
			c.ShowList = false
			return
		}
		c.ShowList = true
		if filter == "" {
			c.List.ResetFilter()
		} else {
			c.List.SetFilterText(filter)
		}
		c.Update(cmd)
		return
	}
	c.ShowList = true
	c.Update(rootCommand)
	c.List.SetFilterText(text)
}

func (c *Commands) IsCommand(text string) bool {
	if !strings.HasPrefix(text, "/") {
		return false
	}
	text = strings.TrimSpace(text[1:])
	for cmd := range c.commands {
		if strings.HasPrefix(cmd, text) {
			return true
		}
	}
	return false
}
