package main

import (
	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

const (
	CommandTypeLogin ginka_ecs_go.CommandType = iota + 1
	CommandTypeAddGold
	CommandTypeRename
)

type LoginCommand struct {
	PlayerId uint64
	Name     string
}

func (c LoginCommand) Type() ginka_ecs_go.CommandType {
	return CommandTypeLogin
}

func (c LoginCommand) EntityId() uint64 {
	return c.PlayerId
}

type AddGoldCommand struct {
	PlayerId uint64
	Amount   int64
}

func (c AddGoldCommand) Type() ginka_ecs_go.CommandType {
	return CommandTypeAddGold
}

func (c AddGoldCommand) EntityId() uint64 {
	return c.PlayerId
}

type RenameCommand struct {
	PlayerId uint64
	Name     string
}

func (c RenameCommand) Type() ginka_ecs_go.CommandType {
	return CommandTypeRename
}

func (c RenameCommand) EntityId() uint64 {
	return c.PlayerId
}
