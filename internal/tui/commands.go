package tui

import (
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type item[T any] struct {
	title, desc string

	extras T
}

func (i item[T]) Title() string       { return i.title }
func (i item[T]) Description() string { return i.desc }
func (i item[T]) FilterValue() string { return i.title }

var commands = map[Commad][]list.Item{
	ListCommands: commandItems,
}

var commandItems = []list.Item{
	item[any]{title: "models", desc: "list of models"},
	item[any]{title: "test", desc: "list of test"},
}

type Commad int

const (
	ListCommands Commad = iota
	ListModels
)

type Commands struct {
	List     list.Model
	current  Commad
	ShowList bool
	Width    int
}

type modelExtras struct {
	ModelID    string
	ProviderID string
}

func parseModels(providers []config.ProviderConfig) []list.Item {
	var items []list.Item
	for _, provider := range providers {
		for _, model := range provider.Models {
			items = append(items, item[modelExtras]{title: provider.Name, desc: model.Name, extras: modelExtras{
				ModelID:    model.ID,
				ProviderID: provider.ID,
			}})
		}
	}
	return items
}

func NewCommands(providers []config.ProviderConfig) *Commands {
	models := parseModels(providers)

	commands[ListModels] = models
	listStyles := newStyles(true, 0)
	list := list.New(commands[ListCommands], itemDelegate{styles: &listStyles}, 5, 10)
	list.SetShowStatusBar(false)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	return &Commands{
		List:    list,
		current: ListCommands,
	}
}

func (c *Commands) Update(command Commad) {
	if c.current == command {
		return
	}

	items, ok := commands[command]
	if !ok {
		return
	}

	c.current = command
	c.List.SetItems(items)
	slog.Debug("update", "current", c.current)
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
	slog.Debug("view", "current", c.current)
	return c.List.View()
}

func (c *Commands) Sync(text string) {
	c.ShowList = false
	if !strings.HasPrefix(text, "/") {
		return
	}
	if after, ok := strings.CutPrefix(text, "/models "); ok {
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
