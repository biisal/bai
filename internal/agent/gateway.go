package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	fantasy "charm.land/fantasy"
	"github.com/biisal/bai/internal/agent/core/instruction"
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/config"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/git"
	audio "github.com/biisal/bai/internal/player"
	broker "github.com/biisal/bai/internal/pubsub"
)

type Gateway struct {
	mu               sync.RWMutex
	broker           broker.Service
	providers        map[string]fantasy.Provider
	db               repo.Querier
	conversation     *repo.Conversation
	AudioPlayer      *audio.AudioPlayer
	skills           []instruction.Skill
	userInstructions []string
	gitRepo          git.GitRepo
	agent            *Agent
}

func NewGateway(
	ctx context.Context,
	db repo.Querier,
	b broker.Service,
	cfg *config.Config,
	audioPlayer *audio.AudioPlayer,
	skillPaths []string,
	gitRepo git.GitRepo,
) (*Gateway, error) {
	providers, err := buildProviders(cfg.Providers)
	if err != nil {
		return nil, fmt.Errorf("failed to build providers: %w", err)
	}
	activeProvider, activeModel, err := resolveProvider(ctx, db, cfg.Providers)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve provider: %w", err)
	}

	skills := instruction.LoadSkills(skillPaths...)
	userInstructions := instruction.ReadAgentMd()

	provider, ok := providers[activeProvider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", activeProvider)
	}

	agent, err := NewFantasyAgent(ctx, NewFantasyAgentParams{
		Model:            activeModel,
		Provider:         provider,
		UserInstructions: []string{userInstructions},
		Skills:           skills,
		AgentTools:       tools.NewTools(b, cfg.PluginsPath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	g := &Gateway{
		db:               db,
		broker:           b,
		providers:        providers,
		AudioPlayer:      audioPlayer,
		skills:           skills,
		userInstructions: []string{userInstructions},
		gitRepo:          gitRepo,
		agent:            agent,
	}
	return g, nil
}

func (g *Gateway) ActiveConversationTitle() string {
	if g.conversation == nil {
		return "bai | Start a new conversation"
	}
	return fmt.Sprintf("bai | %s", g.conversation.Title)
}

func (g *Gateway) Active() (fantasy.Provider, string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.agent.provider, g.agent.model.Model()
}

func (g *Gateway) Providers() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ids := make([]string, 0, len(g.providers))
	for id := range g.providers {
		ids = append(ids, id)
	}
	return ids
}

func (g *Gateway) SetConversation(conversation *repo.Conversation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.conversation = conversation
}

func (g *Gateway) trySavingMsgToDB(partialReasoning, partialText *strings.Builder) {
	var parts []fantasy.MessagePart
	if partialReasoning.Len() > 0 {
		parts = append(parts, fantasy.ReasoningPart{Text: partialReasoning.String()})
	}
	if partialText.Len() > 0 {
		parts = append(parts, fantasy.TextPart{Text: partialText.String()})
	}

	if len(parts) == 0 {
		return
	}

	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if saveErr := g.AddMessageToDB(saveCtx, fantasy.Message{
		Role:    fantasy.MessageRoleAssistant,
		Content: parts,
	}); saveErr != nil {
		slog.Error("failed to save interrupted message", "error", saveErr)
	}
}

// TODO : write test properly
func (g *Gateway) syncGitRepo(purpose string) {
	isDirty, output, err := g.gitRepo.CheckIfDirty()
	if err != nil {
		slog.Error("failed to check if git repo is dirty", "error", err)
		return
	}
	if !isDirty {
		return
	}
	if err := g.gitRepo.Add("."); err != nil {
		slog.Error("failed to add files to git", "error", err)
	}
	msg := strings.TrimSpace(purpose)
	if msg == "" {
		maxLines := 10
		lines := strings.Split(output, "\n")
		linesLen := len(lines)
		if linesLen > maxLines {
			lines = append(lines[:maxLines], fmt.Sprintf("... More %d lines", linesLen-maxLines))
		}

		linesStr := strings.Join(lines, "\n")

		msg = fmt.Sprintf("auto-commit: %s\n%s", time.Now().Format(time.RFC3339), linesStr)
	}
	if err := g.gitRepo.Commit(msg); err != nil {
		slog.Error("failed to commit files", "error", err)
	}
}

func (g *Gateway) StreamChat(ctx context.Context, message string) error {
	defer g.broker.Publish(context.Background(), broker.Message{Type: broker.EventStreamDone, IsComplete: true})
	// 1. Save user message.
	if err := g.AddMessageToDB(ctx, fantasy.Message{
		Role:    fantasy.MessageRoleUser,
		Content: []fantasy.MessagePart{fantasy.TextPart{Text: message}},
	}); err != nil {
		slog.Error("failed to add user message to db", "error", err)
		return err
	}
	g.broker.Publish(ctx, broker.Message{Type: broker.EventUserMessage, Text: message, IsComplete: true})

	// 2. Load conversation history.
	history, err := g.GetMessagesByConversationID(ctx, g.conversation.ID)
	if err != nil {
		slog.Error("failed to get messages", "error", err)
		return err
	}

	var partialReasoning strings.Builder
	var partialText strings.Builder

	var purpose strings.Builder
	g.mu.Lock()
	client := g.agent.client
	g.mu.Unlock()
	_, err = client.Stream(ctx, fantasy.AgentStreamCall{
		Messages: history,
		OnRetry:  fantasy.DefaultRetryOptions().OnRetry,

		OnToolCall: func(toolCall fantasy.ToolCallContent) error {
			slog.Debug("tool call", "input", toolCall.Input, "name", toolCall.ToolName)
			if p := tools.PurposeFromToolCall(toolCall.ToolName, toolCall.Input); p != "" {
				if purpose.Len() > 0 {
					purpose.WriteString("; ")
				}
				purpose.WriteString(p)
			}
			return nil
		},

		OnAgentStart: func() {
			g.broker.Publish(ctx, broker.Message{Type: broker.EventStreamStarted, IsComplete: true})
		},

		OnTextDelta: func(_ string, text string) error {
			partialText.WriteString(text)
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentResponse, Text: text})
			return nil
		},

		OnReasoningDelta: func(_ string, text string) error {
			partialReasoning.WriteString(text)
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentThinking, Text: text})
			return nil
		},

		OnStepFinish: func(step fantasy.StepResult) error {
			partialReasoning.Reset()
			partialText.Reset()
			for _, fm := range step.Messages {
				if saveErr := g.AddMessageToDB(ctx, fm); saveErr != nil {
					slog.Error("failed to save step message", "error", saveErr)
				}
			}
			return nil
		},
	})
	if err != nil {
		g.trySavingMsgToDB(&partialReasoning, &partialText)
		slog.Error("failed to stream chat", "error", err)
		return err
	}

	g.syncGitRepo(purpose.String())

	return nil
}
