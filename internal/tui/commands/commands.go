package commands

import (
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type ListItem[T any] struct {
	Fields T
}

func (i ListItem[T]) Title() string {
	switch v := any(i.Fields).(type) {
	case CommandItem:
		return v.Name
	case ModelList:
		return v.ModelID
	default:
		return ""
	}
}

func (i ListItem[T]) Description() string {
	switch v := any(i.Fields).(type) {
	case CommandItem:
		return v.Description
	case ModelList:
		return v.ModelID
	default:
		return ""
	}
}

func (i ListItem[T]) FilterValue() string {
	switch v := any(i.Fields).(type) {
	case CommandItem:
		return v.Name
	case ModelList:
		return v.ModelID
	default:
		return ""
	}
}

type CommandItem struct {
	Name        string
	Description string
}

var listCommandItems = []list.Item{
	ListItem[CommandItem]{
		Fields: CommandItem{
			Name:        "models",
			Description: "List all available models",
		},
	},
	ListItem[CommandItem]{
		Fields: CommandItem{
			Name:        "test",
			Description: "Test the TUI",
		},
	},
}

type CommandType int

const (
	ListCommands CommandType = iota
	ListModels
)

type Commands struct {
	List     list.Model
	Current  CommandType
	ShowList bool
	Width    int
	commands map[CommandType][]list.Item
}

func NewCommands(providers []config.ProviderConfig) *Commands {
	models := parseModels(providers)
	commands := make(map[CommandType][]list.Item)
	commands[ListModels] = models
	commands[ListCommands] = listCommandItems

	listStyles := newStyles(true, 0)
	list := list.New(commands[ListCommands], itemDelegate{styles: &listStyles}, 5, 10)
	list.SetShowStatusBar(false)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	return &Commands{
		List:     list,
		Current:  ListCommands,
		commands: commands,
	}
}

func (c *Commands) Update(command CommandType) {
	if c.Current == command {
		return
	}

	items, ok := c.commands[command]
	if !ok {
		slog.Warn("update_commands", "command", command)
		return
	}

	c.Current = command
	c.List.SetItems(items)
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
	c.ShowList = false
	if !strings.HasPrefix(text, "/") {
		return
	}
	if after, ok := strings.CutPrefix(text, "/models "); ok { // TODO: change from hardcoded command
		filterCommand := strings.TrimSpace(after)

		c.ShowList = true
		if filterCommand == "" {
			c.List.ResetFilter()
		} else {
			c.List.SetFilterText(filterCommand)
		}
		c.Update(ListModels)
		return
	}
	if strings.HasPrefix(text, "/") && !strings.Contains(text, " ") {
		c.ShowList = true
		filterCommand := text[1:]
		slog.Debug("sync", "filterCommand", filterCommand)
		c.List.SetFilterText(filterCommand)
		c.Update(ListCommands)
	}
}
