package tui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/git"
	broker "github.com/biisal/bai/internal/pubsub"
	chatbuilder "github.com/biisal/bai/internal/tui/chat-builder"
	"github.com/biisal/bai/internal/tui/commands"
)

type chatContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type Model struct {
	gateway         *agent.Gateway
	git             *git.Git
	broker          broker.Service
	messages        <-chan broker.Message
	components      *Component
	Width           int
	Height          int
	ChatContent     *strings.Builder
	ThinkingContent *strings.Builder
	ctx             context.Context
	content         *chatbuilder.Content

	windowTitle string

	chatCtx      *chatContext
	messageQueue messageQueue

	commands *commands.Commands
}

func InitModel(ctx context.Context, gateway *agent.Gateway, broker broker.Service, providers []config.ProviderConfig, gitRepo *git.Git) *Model {
	comp := NewComponent(gitRepo)
	commands := commands.NewCommands(ctx, providers, gateway)

	return &Model{
		gateway:    gateway,
		messages:   broker.Subscribe(),
		components: comp, ctx: ctx,
		ChatContent:     &strings.Builder{},
		ThinkingContent: &strings.Builder{},
		broker:          broker,
		content:         chatbuilder.NewContent(),
		commands:        commands,
		windowTitle:     fmt.Sprintf("bai - %s", files.GetBaseDir()),
		git:             gitRepo,
	}
}

func waitForMsg(msgChan <-chan broker.Message) tea.Cmd {
	return func() tea.Msg {
		return <-msgChan
	}
}

func gitInit(m *Model) tea.Cmd {
	return func() tea.Msg {
		slog.Debug("git init started")

		if m.git == nil {
			slog.Debug("git init", "error", "git repo is nil")
			return nil
		}

		initialized, err := m.git.CheckIfGitInitialized()
		if err != nil {
			slog.Error("git status check", "error", err)
			return err
		}
		if initialized {
			return nil
		}
		settings, err := m.gateway.GetDirectorySettings(m.ctx, files.CurrentDir())
		if err != nil && errors.Is(err, sql.ErrNoRows) || err == nil && settings.AutoGitInit {
			return GitInitMsg{
				showPrompt: true,
			}
		}
		if !settings.AutoGitInit {
			return nil
		}
		return nil
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForMsg(m.messages), gitInit(&m))
}
