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

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

func TestHTTPServerFlow(t *testing.T) {
	baseDir := t.TempDir()
	world := NewGameWorld("http-world")
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

	server := NewServer(world, authSys, walletSys, profileSys)
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

	post("/login", map[string]any{"player_id": "1001", "name": "Aki"})
	post("/add-gold", map[string]any{"player_id": "1001", "amount": 120})
	post("/rename", map[string]any{"player_id": "1001", "name": "AkiHero"})
	post("/login", map[string]any{"player_id": "2002", "name": "Mio"})
	post("/add-gold", map[string]any{"player_id": "2002", "amount": 45})

	checkPlayer := func(id string, name string, gold int64) {
		entity, ok := world.Entities.Get(id)
		if !ok {
			t.Fatalf("expected player %s", id)
		}
		profile, ok := ginka_ecs_go.Get[*ProfileComponent](entity, ComponentTypeProfile)
		if !ok {
			t.Fatalf("expected profile component")
		}
		if profile.Name != name {
			t.Fatalf("profile name = %q", profile.Name)
		}
		wallet, ok := ginka_ecs_go.Get[*WalletComponent](entity, ComponentTypeWallet)
		if !ok {
			t.Fatalf("expected wallet component")
		}
		if wallet.Gold != gold {
			t.Fatalf("wallet gold = %d", wallet.Gold)
		}
	}

	checkPlayer("1001", "AkiHero", 120)
	checkPlayer("2002", "Mio", 45)

	if err := persistenceSys.Flush(context.Background(), world); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, "1001", "profile.json")); err != nil {
		t.Fatalf("expected profile persisted: %v", err)
	}
}
