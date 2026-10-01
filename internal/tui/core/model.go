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
	repo "github.com/biisal/bai/internal/db/sqlc"
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

type directorySettings interface {
	GetDirectorySettings(ctx context.Context, directory string) (repo.DirectorySetting, error)
	UpsertDirectorySettings(ctx context.Context, arg repo.UpsertDirectorySettingsParams) error
}

type Model struct {
	gateway         *agent.Gateway
	settings        directorySettings
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

func InitModel(ctx context.Context, gateway *agent.Gateway, settings directorySettings, broker broker.Service, providers []config.ProviderConfig, gitRepo *git.Git) *Model {
	comp := NewComponent(gitRepo)
	commands := commands.NewCommands(ctx, providers, gateway)

	return &Model{
		gateway:    gateway,
		settings:   settings,
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

		initialized, err := m.git.CheckIfGitInitialized()
		if err != nil {
			slog.Error("git status check", "error", err)
			return err
		}
		if initialized {
			return nil
		}
		settings, err := m.settings.GetDirectorySettings(m.ctx, files.CurrentDir())
		if errors.Is(err, sql.ErrNoRows) || err == nil && settings.AutoGitInit {
			return commands.GitInitMsg{ShowPrompt: true}
		}
		return nil
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForMsg(m.messages), gitInit(&m))
}
