package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

func main() {
	ctx := context.Background()
	world := ginka_ecs_go.NewCoreWorld("demo-world", ginka_ecs_go.WithTickInterval(0))
	if err := world.Register(&AuthSystem{}, &ProfileSystem{}, &WalletSystem{}, NewFilePersistenceSystem("tmp/server_demo")); err != nil {
		log.Fatal(err)
	}
	if err := world.Run(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := world.Stop(); err != nil {
			log.Println(err)
		}
	}()

	playerId := uint64(1001)

	fmt.Println("api: login")
	if err := world.Submit(ctx, LoginCommand{PlayerId: playerId, Name: "Aki"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("api: add gold +120")
	if err := world.Submit(ctx, AddGoldCommand{PlayerId: playerId, Amount: 120}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("api: rename to AkiHero")
	if err := world.Submit(ctx, RenameCommand{PlayerId: playerId, Name: "AkiHero"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("tick: flush dirty components")
	if err := world.TickOnce(ctx, time.Second); err != nil {
		log.Fatal(err)
	}
}
