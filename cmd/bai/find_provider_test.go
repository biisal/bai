package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	test_utils "github.com/biisal/bai/utils/tests"
)

func getTx(t *testing.T, conn *sql.DB) repo.Querier {
	t.Helper()
	tx, err := conn.Begin()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil {
			t.Fatal(err.Error())
		}
	})
	return repo.New(tx)
}

func assertProviderAndModelID(t *testing.T, gotProviderId, gotModelId, wantProviderId, wantModelId string) {
	t.Helper()
	if wantProviderId != gotProviderId {
		t.Errorf("want provider id %s, got %s", wantProviderId, gotProviderId)
	}
	if wantModelId != gotModelId {
		t.Errorf("want model id %s, got %s", wantModelId, gotModelId)
	}
}

func TestFindProvider(t *testing.T) {
	conn, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	tests := []struct {
		name           string
		providers      []config.ProviderConfig
		setup          func(ctx context.Context, t *testing.T) repo.Querier
		wantErr        error
		wantProviderId string
		wantModelId    string
	}{
		{
			wantProviderId: "openai",
			wantModelId:    "gpt-3.5-turbo",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				return getTx(t, conn)
			},
			name: "Return first provider if not found in db",
			providers: []config.ProviderConfig{
				{
					Name: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
					},
				},
			},
		},
		{
			wantProviderId: "openai",
			wantModelId:    "gpt-3.5-turbo",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
					ProviderName: sql.NullString{String: "invalid-id", Valid: true},
					ModelID:      sql.NullString{String: "gpt-3.5-turbo", Valid: true},
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			name: "Return first provider and model if db record does not exist in config",
			providers: []config.ProviderConfig{
				{
					Name: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
					},
				},
			},
		},
		{
			wantProviderId: "openai",
			wantModelId:    "gpt-3.5-turbo",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
					ProviderName: sql.NullString{String: "openai", Valid: true},
					ModelID:      sql.NullString{String: "test-model", Valid: true},
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			name: "Return first provider and model if db has correct provider but model does not exist in config",
			providers: []config.ProviderConfig{
				{
					Name: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
					},
				},
			},
		},
		{
			name:           "Return db provider and model if both provider and model available in config",
			wantProviderId: "openai",
			wantModelId:    "test-model-available",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
					ProviderName: sql.NullString{String: "openai", Valid: true},
					ModelID:      sql.NullString{String: "test-model-available", Valid: true},
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			providers: []config.ProviderConfig{
				{
					Name: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
						{
							ID: "test-model-available",
						},
					},
				},
			},
		},
		{
			name:           "returns matched db provider and model when multiple providers",
			wantProviderId: "openai",
			wantModelId:    "test-model-available",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
					ProviderName: sql.NullString{String: "openai", Valid: true},
					ModelID:      sql.NullString{String: "test-model-available", Valid: true},
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			providers: []config.ProviderConfig{
				{
					Name: "grok",
					Models: []config.ModelConfig{
						{
							ID: "some-model",
						},
					},
				},
				{
					Name: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
						{
							ID: "test-model-available",
						},
					},
				},
			},
		},
		{
			name:    "returns error when providers list is empty",
			wantErr: ErrProviderNotFound,
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				return getTx(t, conn)
			},
			providers: []config.ProviderConfig{},
		},
		{
			name:    "returns error when first provider has no models",
			wantErr: ErrModelNotFound,
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				return getTx(t, conn)
			},
			providers: []config.ProviderConfig{
				{
					Name:   "openai",
					Models: []config.ModelConfig{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := tt.setup(ctx, t)
			gotProvider, gotModelId, err := resolveProvider(ctx, tx, tt.providers)
			test_utils.AssertError(t, err, tt.wantErr)

			if err != nil {
				t.Log(err.Error())
			}
			assertProviderAndModelID(t, gotProvider, gotModelId, tt.wantProviderId, tt.wantModelId)
		})
	}
}
