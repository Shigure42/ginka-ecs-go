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

func (s *AuthSystem) SubscribedCommands() []ginka_ecs_go.CommandType {
	return []ginka_ecs_go.CommandType{CommandTypeLogin}
}

func (s *AuthSystem) Handle(ctx context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
	login, ok := cmd.(LoginCommand)
	if !ok {
		return fmt.Errorf("auth system: unexpected command %T", cmd)
	}
	if _, exists := w.Entities().Get(login.PlayerId); exists {
		return nil
	}
	player, err := w.Entities().Create(ctx, login.PlayerId, login.Name, EntityTypePlayer, TagPlayer)
	if err != nil {
		return err
	}
	if err := player.SetData(NewProfileComponent(login.Name)); err != nil {
		return err
	}
	if err := player.SetData(NewWalletComponent(0)); err != nil {
		return err
	}
	return nil
}

type WalletSystem struct{}

func (s *WalletSystem) Name() string {
	return "wallet"
}

func (s *WalletSystem) SubscribedCommands() []ginka_ecs_go.CommandType {
	return []ginka_ecs_go.CommandType{CommandTypeAddGold}
}

func (s *WalletSystem) Handle(_ context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
	addGold, ok := cmd.(AddGoldCommand)
	if !ok {
		return fmt.Errorf("wallet system: unexpected command %T", cmd)
	}
	player, exists := w.Entities().Get(addGold.PlayerId)
	if !exists {
		return fmt.Errorf("wallet system: player %d: %w", addGold.PlayerId, ginka_ecs_go.ErrEntityNotFound)
	}
	return player.MutateData(ComponentTypeWallet, func(c ginka_ecs_go.DataComponent) error {
		wallet, ok := c.(*WalletComponent)
		if !ok {
			return fmt.Errorf("wallet system: component %T", c)
		}
		wallet.Gold += addGold.Amount
		return nil
	})
}

type ProfileSystem struct{}

func (s *ProfileSystem) Name() string {
	return "profile"
}

func (s *ProfileSystem) SubscribedCommands() []ginka_ecs_go.CommandType {
	return []ginka_ecs_go.CommandType{CommandTypeRename}
}

func (s *ProfileSystem) Handle(_ context.Context, w ginka_ecs_go.World, cmd ginka_ecs_go.Command) error {
	rename, ok := cmd.(RenameCommand)
	if !ok {
		return fmt.Errorf("profile system: unexpected command %T", cmd)
	}
	player, exists := w.Entities().Get(rename.PlayerId)
	if !exists {
		return fmt.Errorf("profile system: player %d: %w", rename.PlayerId, ginka_ecs_go.ErrEntityNotFound)
	}
	return player.MutateData(ComponentTypeProfile, func(c ginka_ecs_go.DataComponent) error {
		profile, ok := c.(*ProfileComponent)
		if !ok {
			return fmt.Errorf("profile system: component %T", c)
		}
		profile.Name = rename.Name
		return nil
	})
}
