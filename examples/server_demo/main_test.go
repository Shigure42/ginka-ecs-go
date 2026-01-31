package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

func TestServerDemoFlow(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()
	world := ginka_ecs_go.NewCoreWorld("test-world")
	if err := world.Register(&AuthSystem{}, &ProfileSystem{}, &WalletSystem{}, NewFilePersistenceSystem(baseDir)); err != nil {
		t.Fatalf("register systems: %v", err)
	}
	if err := world.Run(); err != nil {
		t.Fatalf("run world: %v", err)
	}
	defer func() {
		if err := world.Stop(); err != nil {
			t.Fatalf("stop world: %v", err)
		}
	}()

	playerId := uint64(1001)

	if err := world.Submit(ctx, ginka_ecs_go.NewAction(LoginCommand{PlayerId: playerId, Name: "Aki"})); err != nil {
		t.Fatalf("submit login: %v", err)
	}
	if err := world.Submit(ctx, ginka_ecs_go.NewAction(AddGoldCommand{PlayerId: playerId, Amount: 120})); err != nil {
		t.Fatalf("submit add gold: %v", err)
	}
	if err := world.Submit(ctx, ginka_ecs_go.NewAction(RenameCommand{PlayerId: playerId, Name: "AkiHero"})); err != nil {
		t.Fatalf("submit rename: %v", err)
	}

	entity, ok := world.Entities().Get(playerId)
	if !ok {
		t.Fatalf("expected player %d", playerId)
	}
	profileData, ok := entity.GetData(ComponentTypeProfile)
	if !ok {
		t.Fatalf("expected profile component")
	}
	profile, ok := profileData.(*ProfileComponent)
	if !ok {
		t.Fatalf("profile component type mismatch")
	}
	if profile.Name != "AkiHero" {
		t.Fatalf("profile name = %q", profile.Name)
	}
	walletData, ok := entity.GetData(ComponentTypeWallet)
	if !ok {
		t.Fatalf("expected wallet component")
	}
	wallet, ok := walletData.(*WalletComponent)
	if !ok {
		t.Fatalf("wallet component type mismatch")
	}
	if wallet.Gold != 120 {
		t.Fatalf("wallet gold = %d", wallet.Gold)
	}

	if err := world.Submit(ctx, ginka_ecs_go.NewTick(time.Second)); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if len(entity.DirtyTypes()) != 0 {
		t.Fatalf("expected dirty types cleared")
	}

	profilePath := filepath.Join(baseDir, "1001", "profile.json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	var profileDisk struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &profileDisk); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	if profileDisk.Name != "AkiHero" {
		t.Fatalf("profile disk name = %q", profileDisk.Name)
	}

	walletPath := filepath.Join(baseDir, "1001", "wallet.json")
	data, err = os.ReadFile(walletPath)
	if err != nil {
		t.Fatalf("read wallet: %v", err)
	}
	var walletDisk struct {
		Gold int64 `json:"gold"`
	}
	if err := json.Unmarshal(data, &walletDisk); err != nil {
		t.Fatalf("unmarshal wallet: %v", err)
	}
	if walletDisk.Gold != 120 {
		t.Fatalf("wallet disk gold = %d", walletDisk.Gold)
	}
}
