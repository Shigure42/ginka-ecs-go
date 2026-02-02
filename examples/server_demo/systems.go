package main

import (
	"context"
	"fmt"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

type AuthSystem struct{}

func (s *AuthSystem) Name() string {
	return "auth"
}

func (s *AuthSystem) Login(ctx context.Context, w *GameWorld, login LoginRequest) error {
	if _, exists := w.Entities.Get(login.PlayerId); exists {
		return nil
	}
	player, err := w.Entities.Create(ctx, login.PlayerId, login.Name, EntityTypePlayer, TagPlayer)
	if err != nil {
		return err
	}
	return player.Tx(func(tx ginka_ecs_go.DataEntity) error {
		if err := tx.Add(NewProfileComponent(login.Name)); err != nil {
			return err
		}
		if err := tx.Add(NewWalletComponent(0)); err != nil {
			return err
		}
		// Mark new components as dirty so they will be persisted.
		tx.GetForUpdate(ComponentTypeProfile)
		tx.GetForUpdate(ComponentTypeWallet)
		return nil
	})
}

type WalletSystem struct{}

func (s *WalletSystem) Name() string {
	return "wallet"
}

func (s *WalletSystem) AddGold(_ context.Context, w *GameWorld, addGold AddGoldRequest) error {
	player, exists := w.Entities.Get(addGold.PlayerId)
	if !exists {
		return fmt.Errorf("wallet system: player %s: %w", addGold.PlayerId, ginka_ecs_go.ErrEntityNotFound)
	}
	return player.Tx(func(tx ginka_ecs_go.DataEntity) error {
		wallet, ok := ginka_ecs_go.GetForUpdate[*WalletComponent](tx, ComponentTypeWallet)
		if !ok {
			return fmt.Errorf("wallet system: component %d: %w", ComponentTypeWallet, ginka_ecs_go.ErrComponentNotFound)
		}
		wallet.Gold += addGold.Amount
		return nil
	})
}

type ProfileSystem struct{}

func (s *ProfileSystem) Name() string {
	return "profile"
}

func (s *ProfileSystem) Rename(_ context.Context, w *GameWorld, rename RenameRequest) error {
	player, exists := w.Entities.Get(rename.PlayerId)
	if !exists {
		return fmt.Errorf("profile system: player %s: %w", rename.PlayerId, ginka_ecs_go.ErrEntityNotFound)
	}
	return player.Tx(func(tx ginka_ecs_go.DataEntity) error {
		profile, ok := ginka_ecs_go.GetForUpdate[*ProfileComponent](tx, ComponentTypeProfile)
		if !ok {
			return fmt.Errorf("profile system: component %d: %w", ComponentTypeProfile, ginka_ecs_go.ErrComponentNotFound)
		}
		profile.Name = rename.Name
		return nil
	})
}
