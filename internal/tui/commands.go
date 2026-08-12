package tui

import (
	"charm.land/bubbles/v2/list"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var commandItems = []list.Item{
	item{title: "models", desc: "list of models"},
	item{title: "test", desc: "list of test"},
}

var modelItems = []list.Item{
	item{title: "models ", desc: "list of models"},
	item{title: "test", desc: "list of test"},
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

func NewCommands() *Commands {
	return &Commands{
		List:    list.New(commandItems, list.NewDefaultDelegate(), 5, 1),
		current: ListCommands,
	}
}

func (c *Commands) Update(command Commad) {
	if c.current == command {
		return
	}

	items := commandItems
	switch command {
	case ListCommands:
		items = commandItems
	case ListModels:
		items = modelItems
	}

	c.current = command
	c.List.SetItems(items)
}

func (c *Commands) SetSize(width int) {
	c.Width = width
}

func (c *Commands) View() string {
	if !c.ShowList {
		return ""
	}
	c.List.SetWidth(c.Width)
	return c.List.View()
}
