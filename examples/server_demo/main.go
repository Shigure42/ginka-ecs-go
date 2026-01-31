package main

import (
	"context"
	"fmt"
	"log"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

func main() {
	ctx := context.Background()
	world := ginka_ecs_go.NewCoreWorld("demo-world")
	authSys := &AuthSystem{}
	profileSys := &ProfileSystem{}
	walletSys := &WalletSystem{}
	persistenceSys := NewFilePersistenceSystem("tmp/server_demo")
	if err := world.Register(authSys, profileSys, walletSys, persistenceSys); err != nil {
		log.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- world.Run()
	}()
	defer func() {
		if err := world.Stop(); err != nil {
			log.Println(err)
		}
		if err := <-runDone; err != nil {
			log.Println(err)
		}
	}()

	playerId := uint64(1001)

	fmt.Println("api: login")
	if err := authSys.Login(ctx, world, LoginRequest{PlayerId: playerId, Name: "Aki"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("api: add gold +120")
	if err := walletSys.AddGold(ctx, world, AddGoldRequest{PlayerId: playerId, Amount: 120}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("api: rename to AkiHero")
	if err := profileSys.Rename(ctx, world, RenameRequest{PlayerId: playerId, Name: "AkiHero"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("flush: dirty components")
	if err := persistenceSys.Flush(ctx, world); err != nil {
		log.Fatal(err)
	}
}
