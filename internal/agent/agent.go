package agent

import (
	"context"
	"database/sql"
	"fmt"

	"charm.land/fantasy"
	"github.com/biisal/bai/internal/agent/core/instruction"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/skills"
)

type NewFantasyAgentParams struct {
	Model            string
	Provider         fantasy.Provider
	UserInstructions []string
	Skills           []skills.Skill
	AgentTools       []fantasy.AgentTool
}

type Agent struct {
	model            fantasy.LanguageModel
	provider         fantasy.Provider
	client           fantasy.Agent
	userInstructions []string
	skills           []skills.Skill
	agentTools       []fantasy.AgentTool
}

func NewFantasyAgent(ctx context.Context, params NewFantasyAgentParams) (*Agent, error) {
	model, err := params.Provider.LanguageModel(ctx, params.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get language model: %w", err)
	}

	ag := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(instruction.BuildSystemPrompt(params.UserInstructions, params.Skills, params.AgentTools)),
		fantasy.WithTools(params.AgentTools...),
		fantasy.WithMaxRetries(3),
	)

	return &Agent{
		client:           ag,
		userInstructions: params.UserInstructions,
		skills:           params.Skills,
		agentTools:       params.AgentTools,
		model:            model,
		provider:         params.Provider,
	}, nil
}

func (g *Gateway) AddOrUpdateProvider(ctx context.Context, providerID, modelId string) error {
	g.mu.Lock()
	provider, ok := g.providers[providerID]
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("unknown provider: %s", providerID)
	}

	model, err := provider.LanguageModel(ctx, modelId)
	if err != nil {
		return fmt.Errorf("failed to get language model: %w", err)
	}

	if err := g.db.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
		ProviderName: sql.NullString{Valid: true, String: providerID},
		ModelID:      sql.NullString{Valid: true, String: modelId},
	}); err != nil {
		return err
	}
	ag := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(instruction.BuildSystemPrompt(g.agent.userInstructions, g.agent.skills, g.agent.agentTools)),
		fantasy.WithTools(g.agent.agentTools...),
		fantasy.WithMaxRetries(3),
	)

	g.mu.Lock()
	defer g.mu.Unlock()
	g.agent.client = ag
	g.agent.model = model
	g.agent.provider = provider

	return nil
}
