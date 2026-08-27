package commands

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"charm.land/fantasy"
	broker "github.com/biisal/bai/internal/pubsub"
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
	commands map[string]*commandEntry
	gateway  *agent.Gateway
	ctx      context.Context

	models    []list.Item
	rootItems []list.Item

	lastSynced string
}

type CommandContext struct {
	Gateway    *agent.Gateway
	Broker     broker.Service
	Content    Content
	Components Components
	ShowList   *bool
}

type Content interface {
	ReRenderFromDbConversation(messages []fantasy.Message)
	Render() string
}

type Components interface {
	SetChatContent(content string)
	ScrollChatToBottom()
	SetValue(value string)
}

type commandEntry struct {
	desc string
	fn   func(ctx CommandContext) tea.Cmd
}

func NewCommands(ctx context.Context, providers []config.ProviderConfig, gateway *agent.Gateway) *Commands {
	models := parseModels(providers)
	commands := map[string]*commandEntry{
		"": {
			desc: "root",
			fn:   nil,
		},
		"models": {
			desc: "show available models",
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
		},
		"sessions": {
			desc: "show list of conversations",
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
		},
		"exit": {
			desc: "exit the application",
			fn: func(c CommandContext) tea.Cmd {
				*c.ShowList = false
				c.Broker.Publish(ctx, broker.Message{
					Type:       broker.EventSystemNotice,
					Text:       "Bye.. See you soon!\n",
					IsComplete: true,
				})
				return func() tea.Msg {
					return tea.Quit()
				}
			},
		},
		"new": {
			desc: "create a new conversation",
			fn: func(c CommandContext) tea.Cmd {
				c.Gateway.SetConversation(nil)
				c.Components.SetValue("")
				c.Content.ReRenderFromDbConversation(nil)
				c.Broker.Publish(ctx, broker.Message{
					Type:       broker.EventSystemNotice,
					Text:       "New conversation started.",
					IsComplete: true,
				})
				*c.ShowList = false
				return nil
			},
		},
	}

	listStyles := newStyles(0)

	rootItems := make([]list.Item, 0)
	for name, entry := range commands {
		if name == "" {
			continue
		}
		rootItems = append(rootItems, CommandItem{Name: name, Desc: entry.desc})
	}

	commands[""].fn = nil

	l := list.New(rootItems, itemDelegate{styles: &listStyles}, 5, 10)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)

	l.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("ctrl+p", "up", "shift+tab")),
		CursorDown: key.NewBinding(key.WithKeys("ctrl+n", "down", "tab")),
	}
	return &Commands{
		List:      l,
		Current:   rootCommand,
		commands:  commands,
		gateway:   gateway,
		ctx:       ctx,
		models:    models,
		rootItems: rootItems,
	}
}

func (c *Commands) ExecuteCommand(command string, cmdCtx CommandContext) tea.Cmd {
	entry, ok := c.commands[command]
	if !ok {
		return nil
	}

	if entry.fn != nil {
		return entry.fn(cmdCtx)
	}
	return nil
}

func (c *Commands) getItems(command string) []list.Item {
	switch command {
	case "models":
		return c.models
	case "sessions":
		return toListItems(parseConversations(c.ctx, c.gateway.GetConversationsByCurrentDir))
	default:
		return c.rootItems
	}
}

func (c *Commands) SetSize(width int) {
	c.Width = width
	listStyles := newStyles(width)
	d := itemDelegate{styles: &listStyles}
	c.List.SetDelegate(d)
}

func (c *Commands) View() string {
	if !c.ShowList {
		return ""
	}
	c.List.SetWidth(c.Width)
	return c.List.View()
}

func (c *Commands) Sync(text string) {
	if !strings.HasPrefix(text, "/") {
		c.ShowList = false
		c.lastSynced = ""
		return
	}

	if c.lastSynced == text && c.ShowList {
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
		c.Current = cmd
		c.List.SetItems(c.getItems(cmd))
		if filter == "" {
			c.List.ResetFilter()
		} else {
			c.List.SetFilterText(filter)
		}
		return
	}
	c.ShowList = true
	c.Current = rootCommand
	c.List.SetItems(c.rootItems)
	c.List.SetFilterText(text)
}

func (c *Commands) IsCommand(text string) bool {
	text = strings.TrimSpace(text)
	if text == "/" {
		return true
	}
	if !strings.HasPrefix(text, "/") {
		return false
	}

	text = strings.Fields(text[1:])[0]

	for cmd := range c.commands {
		if strings.HasPrefix(cmd, text) {
			return true
		}
	}
	return false
}
