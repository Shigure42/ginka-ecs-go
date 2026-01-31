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
	authSys := &AuthSystem{}
	profileSys := &ProfileSystem{}
	walletSys := &WalletSystem{}
	persistenceSys := NewFilePersistenceSystem(baseDir)
	if err := world.Register(authSys, profileSys, walletSys, persistenceSys); err != nil {
		t.Fatalf("register systems: %v", err)
	}
	runDone := startWorld(t, world)
	defer func() {
		if err := world.Stop(); err != nil {
			t.Fatalf("stop world: %v", err)
		}
		if err := <-runDone; err != nil {
			t.Fatalf("run world: %v", err)
		}
	}()

	playerId := uint64(1001)

	if err := authSys.Login(ctx, world, LoginRequest{PlayerId: playerId, Name: "Aki"}); err != nil {
		t.Fatalf("login: %v", err)
	}
	if err := walletSys.AddGold(ctx, world, AddGoldRequest{PlayerId: playerId, Amount: 120}); err != nil {
		t.Fatalf("add gold: %v", err)
	}
	if err := profileSys.Rename(ctx, world, RenameRequest{PlayerId: playerId, Name: "AkiHero"}); err != nil {
		t.Fatalf("rename: %v", err)
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

	if err := persistenceSys.Flush(ctx, world); err != nil {
		t.Fatalf("flush: %v", err)
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

func startWorld(t *testing.T, world ginka_ecs_go.World) chan error {
	t.Helper()
	runDone := make(chan error, 1)
	go func() {
		runDone <- world.Run()
	}()
	waitForRunning(t, world)
	return runDone
}

func waitForRunning(t *testing.T, world ginka_ecs_go.World) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		if world.IsRunning() {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("world did not start")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
