package commands

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/fantasy"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/files"
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

var (
	rootCommand = "/"
	fileCommand = "@"
)

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
	ScrollChatToBottom(msg ...broker.Message)
	SetValue(value string)
}

type commandEntry struct {
	desc  string
	fn    func(ctx CommandContext) tea.Cmd
	items func(c *Commands) []list.Item
	// matchFn           func(text string) bool
	showResultOnSpace bool
}

func NewCommands(ctx context.Context, providers []config.ProviderConfig,
	gateway *agent.Gateway,
) *Commands {
	models := parseModels(providers)
	commands := map[string]*commandEntry{
		"/": {
			desc: "root",
			fn:   nil,
		},
		"/models": {
			desc:              "show available models",
			showResultOnSpace: true,
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
			items: func(c *Commands) []list.Item {
				return c.models
			},
		},
		"/sessions": {
			desc:              "show list of conversations",
			showResultOnSpace: true,
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
			items: func(c *Commands) []list.Item {
				return toListItems(parseConversations(c.ctx,
					c.gateway.GetConversationsByCurrentDir))
			},
		},
		"/new": {
			desc:              "create a new conversation",
			showResultOnSpace: true,
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
		"/themes": {
			desc:              "list available themes",
			showResultOnSpace: true,
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
			items: func(c *Commands) []list.Item {
				return ThemeFiles()
			},
		},
		"/exit": {
			desc:              "exit the application",
			showResultOnSpace: true,
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
		fileCommand: {
			desc:              "find files",
			showResultOnSpace: false,
			items: func(c *Commands) []list.Item {
				return FileItems()
			},
			fn: func(c CommandContext) tea.Cmd {
				return nil
			},
		},
	}

	listStyles := newStyles(0)

	rootItems := make([]list.Item, 0)
	for name, entry := range commands {
		if name == rootCommand || name == fileCommand {
			continue
		}
		rootItems = append(rootItems, CommandItem{Name: name, Desc: entry.desc})
	}

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

func (c *Commands) ExecuteCommand(command string,
	cmdCtx CommandContext,
) tea.Cmd {
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
	if entry, ok := c.commands[command]; ok && entry.items != nil {
		return entry.items(c)
	}
	return c.rootItems
}

func (c *Commands) HasSubItems(command string) bool {
	entry, ok := c.commands[command]
	return ok && entry.items != nil
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
	if text == c.lastSynced {
		return
	}
	c.lastSynced = text

	command, filter, ok := c.match(text)
	if !ok {
		c.ShowList = false
		return
	}

	c.ShowList = true
	c.Current = command
	c.List.SetItems(c.getItems(command))

	if filter == "" {
		c.List.ResetFilter()
	} else {
		c.List.SetFilterText(filter)
	}
}

func (c *Commands) match(text string) (command, filter string, ok bool) {
	if matched, query := files.IfFileFinding(text); matched {
		return fileCommand, query, true
	}

	if !strings.HasPrefix(text, rootCommand) {
		return "", "", false
	}

	if command, filter, found := strings.Cut(text, " "); found {
		if _, exists := c.commands[command]; exists {
			return command, filter, true
		}
	}

	return rootCommand, strings.TrimPrefix(text, rootCommand), true
}

func (c *Commands) IsCommand(text string) bool {
	text = strings.TrimSpace(text)
	for cmd := range c.commands {
		if strings.HasPrefix(text, cmd) {
			return true
		}
	}

	matched, _ := files.IfFileFinding(text)
	return matched
}
