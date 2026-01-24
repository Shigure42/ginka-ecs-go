package main

import (
	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

const (
	EntityTypePlayer ginka_ecs_go.EntityType = iota + 1
)

const TagPlayer ginka_ecs_go.Tag = "player"

type PlayerEntity struct {
	*ginka_ecs_go.DataEntityCore
}

func NewPlayerEntity(id uint64, name string, typ ginka_ecs_go.EntityType, tags ...ginka_ecs_go.Tag) *PlayerEntity {
	return &PlayerEntity{
		DataEntityCore: ginka_ecs_go.NewDataEntityCore(id, name, typ, tags...),
	}
}
