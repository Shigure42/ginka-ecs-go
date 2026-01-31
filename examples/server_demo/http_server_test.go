package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

func TestHTTPServerFlow(t *testing.T) {
	baseDir := t.TempDir()
	world := ginka_ecs_go.NewCoreWorld("http-world")
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

	server := NewServer(world)
	httpServer := httptest.NewServer(server.Routes())
	defer httpServer.Close()

	post := func(path string, payload any) {
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		resp, err := http.Post(httpServer.URL+path, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("post %s status = %d", path, resp.StatusCode)
		}
	}

	post("/login", map[string]any{"player_id": 1001, "name": "Aki"})
	post("/add-gold", map[string]any{"player_id": 1001, "amount": 120})
	post("/rename", map[string]any{"player_id": 1001, "name": "AkiHero"})
	post("/login", map[string]any{"player_id": 2002, "name": "Mio"})
	post("/add-gold", map[string]any{"player_id": 2002, "amount": 45})

	checkPlayer := func(id uint64, name string, gold int64) {
		entity, ok := world.Entities().Get(id)
		if !ok {
			t.Fatalf("expected player %d", id)
		}
		profileData, ok := entity.GetData(ComponentTypeProfile)
		if !ok {
			t.Fatalf("expected profile component")
		}
		profile, ok := profileData.(*ProfileComponent)
		if !ok {
			t.Fatalf("profile component type mismatch")
		}
		if profile.Name != name {
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
		if wallet.Gold != gold {
			t.Fatalf("wallet gold = %d", wallet.Gold)
		}
	}

	checkPlayer(1001, "AkiHero", 120)
	checkPlayer(2002, "Mio", 45)

	if err := world.Submit(context.Background(), ginka_ecs_go.NewTick(time.Second)); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, "1001", "profile.json")); err != nil {
		t.Fatalf("expected profile persisted: %v", err)
	}
}
