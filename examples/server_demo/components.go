package main

import (
	"encoding/json"
	"fmt"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

const (
	ComponentTypeProfile ginka_ecs_go.ComponentType = iota + 1
	ComponentTypeWallet
)

type ProfileComponent struct {
	ginka_ecs_go.DataComponentCore
	Name string `json:"name"`
}

func NewProfileComponent(name string) *ProfileComponent {
	return &ProfileComponent{
		DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypeProfile),
		Name:              name,
	}
}

func (c *ProfileComponent) StorageKey() string {
	return "profile"
}

func (c *ProfileComponent) Marshal() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshal profile: %w", err)
	}
	return data, nil
}

func (c *ProfileComponent) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal profile: %w", err)
	}
	return nil
}

type WalletComponent struct {
	ginka_ecs_go.DataComponentCore
	Gold int64 `json:"gold"`
}

func NewWalletComponent(gold int64) *WalletComponent {
	return &WalletComponent{
		DataComponentCore: ginka_ecs_go.NewDataComponentCore(ComponentTypeWallet),
		Gold:              gold,
	}
}

func (c *WalletComponent) StorageKey() string {
	return "wallet"
}

func (c *WalletComponent) Marshal() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshal wallet: %w", err)
	}
	return data, nil
}

func (c *WalletComponent) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal wallet: %w", err)
	}
	return nil
}
