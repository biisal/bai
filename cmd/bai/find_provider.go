package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
)

func getOrSetProvider(ctx context.Context, svc db.ServiceInterface, providers []config.ProviderConfig) (providerID, modelID string, err error) {
	if len(providers) == 0 {
		return "", "", errors.New("no providers configured")
	}
	if len(providers[0].Models) == 0 {
		return "", "", errors.New("no models configured")
	}
	name := providers[0].Name
	modelId := providers[0].Models[0].ID
	providerId := providers[0].ID
	if err := svc.AddOrUpdateProvider(ctx, name, providerId, modelId); err != nil {
		return "", "", err
	}
	return providerId, modelId, nil
}

func resolveProvider(ctx context.Context, svc db.ServiceInterface, providers []config.ProviderConfig) (providerID, modelID string, err error) {
	provider, err := svc.GetProvider(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return getOrSetProvider(ctx, svc, providers)
		}
		return "", "", err
	}
	for _, p := range providers {
		if p.ID == provider.ProviderID.String {
			for _, model := range p.Models {
				if model.ID == provider.ModelID.String {
					return p.ID, model.ID, nil
				}
			}
			if len(p.Models) == 0 {
				return getOrSetProvider(ctx, svc, providers)
			}
			return p.ID, p.Models[0].ID, nil
		}
	}
	return getOrSetProvider(ctx, svc, providers)
}
