package main

import ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"

type GameWorld struct {
	*ginka_ecs_go.CoreWorld
	Entities ginka_ecs_go.EntityManager[ginka_ecs_go.DataEntity]
}

func NewGameWorld(name string) *GameWorld {
	entities := ginka_ecs_go.NewEntityManager(func(id string, entityName string, typ ginka_ecs_go.EntityType, tags ...ginka_ecs_go.Tag) (ginka_ecs_go.DataEntity, error) {
		return ginka_ecs_go.NewDataEntityCore(id, entityName, typ, tags...), nil
	}, 0)
	return &GameWorld{
		CoreWorld: ginka_ecs_go.NewCoreWorld(name),
		Entities:  entities,
	}
}
