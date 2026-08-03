package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	test_utils "github.com/biisal/bai/utils/tests"
)

func getTx(t *testing.T, conn *sql.DB, dbService db.ServiceInterface) db.ServiceInterface {
	t.Helper()
	tx, err := conn.Begin()
	if err != nil {
		t.Fatal(err.Error())
	}
	// TODO : create all table

	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil {
			t.Fatal(err.Error())
		}
	})
	return dbService.WithTx(tx)
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
	dbService := db.New(conn)
	tests := []struct {
		name           string
		providers      []config.ProviderConfig
		setup          func(t *testing.T) db.ServiceInterface
		wantErr        error
		wantProviderId string
		wantModelId    string
	}{
		{
			wantProviderId: "openai",
			wantModelId:    "gpt-3.5-turbo",
			setup: func(t *testing.T) db.ServiceInterface {
				t.Helper()
				return getTx(t, conn, dbService)
			},
			name: "Return first provider if not found in db",
			providers: []config.ProviderConfig{
				{
					ID: "openai",
					Models: []config.ModelConfig{
						{
							ID: "gpt-3.5-turbo",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			qtx := tt.setup(t)
			gotProvider, gotModelId, err := resolveProvider(ctx, qtx, tt.providers)
			test_utils.AssertError(t, err, tt.wantErr)

			t.Log(err.Error())

			assertProviderAndModelID(t, gotProvider, gotModelId, tt.wantProviderId, tt.wantModelId)
		})
	}
}
